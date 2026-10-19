package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthService struct {
	Repository *AuthRepository

	otpSecret string
	jwtKey    string
	sender    OTPSender
}

func NewAuthService(repo *AuthRepository, secretOTP, jwtKey string, sender OTPSender) *AuthService {
	return &AuthService{
		Repository: repo,
		otpSecret:  secretOTP,
		jwtKey:     jwtKey,
		sender:     sender,
	}
}

const (
	otpTTL         = 5 * time.Minute
	maxOTPAttempts = 5
)

// Generates code and challenge id
func (s *AuthService) StartAuth(
	ctx context.Context,
	rawPhone string,
) (*SafeAuthChallenge, error) {

	phone, err := validatePhoneNumber(rawPhone)
	if err != nil {
		return nil, err
	}

	code := GenOTP()
	id := uuid.New().String()

	now := time.Now()

	challenge := &AuthChallenge{
		ID:          id,
		PhoneNumber: phone,
		CodeHash:    hex.EncodeToString(HashOTP(s.otpSecret, id, code)),
		ExpiresAt:   now.Add(otpTTL),
		Attempts:    0,
		CreatedAt:   now,
	}

	if err := s.Repository.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	if err := s.sender.SendCode(ctx, phone, code); err != nil {
		if delErr := s.Repository.DeleteChallenge(ctx, id); delErr != nil {
			return nil, fmt.Errorf(
				"send OTP: %w; cleanup challenge: %v",
				err,
				delErr,
			)
		}

		return nil, fmt.Errorf("send OTP: %w", err)
	}

	return &SafeAuthChallenge{
		ID:        challenge.ID,
		ExpiresAt: challenge.ExpiresAt,
	}, nil
}

func (s *AuthService) VerifyCode(
	ctx context.Context,
	challengeID string,
	code string,
) (string, error) {
	challenge, err := s.Repository.GetChallenge(ctx, challengeID)
	if err != nil {
		return "", err
	}

	if challenge.Attempts >= maxOTPAttempts {
		if err := s.Repository.DeleteChallenge(ctx, challengeID); err != nil {
			return "", fmt.Errorf("delete exhausted challenge: %w", err)
		}
		return "", ErrTooManyAttempts
	}

	expected := HashOTP(s.otpSecret, challengeID, code)
	actual, err := hex.DecodeString(challenge.CodeHash)
	if err != nil {
		return "", err
	}

	if !hmac.Equal(expected, actual) {
		challenge.Attempts++

		if challenge.Attempts >= maxOTPAttempts {
			if err := s.Repository.DeleteChallenge(ctx, challengeID); err != nil {
				return "", fmt.Errorf("delete exhausted challenge: %w", err)
			}
			return "", ErrTooManyAttempts
		}

		if err := s.Repository.UpdateChallenge(ctx, challenge); err != nil {
			return "", fmt.Errorf("update challenge attempts: %w", err)
		}

		return "", ErrInvalidOTP
	}

	if err := s.Repository.DeleteChallenge(ctx, challengeID); err != nil {
		return "", fmt.Errorf("consume challenge: %w", err)
	}

	return challenge.PhoneNumber, nil
}

func (s *AuthService) CompleteVerification(
	ctx context.Context,
	phone string,
) (*VerificationResult, error) {
	id, err := s.Repository.GetByPhone(ctx, phone)
	if err != nil {
		return &VerificationResult{
			Status: "error",
		}, err
	}

	if id == "" {
		token, err := GenerateToken()
		if err != nil {
			return &VerificationResult{}, err
		}

		err = s.Repository.CreateRegistration(ctx, token, phone)
		if err != nil {
			return &VerificationResult{}, err
		}

		return &VerificationResult{
			Status: "register",
			Token:  token,
		}, nil
	}

	return &VerificationResult{
		Status: "enter_pin",
	}, nil
}

func (s *AuthService) ConfirmCode(
	ctx context.Context,
	challengeID string,
	code string,
) (*VerificationResult, error) {
	phone, err := s.VerifyCode(ctx, challengeID, code)
	if err != nil {
		return nil, err
	}

	return s.CompleteVerification(ctx, phone)
}

func (s *AuthService) Register(
	ctx context.Context,
	regToken string,
	owner string,
	pin string,
	currency string,
) (*TokenPair, error) {
	phone, err := s.Repository.GetRegistration(ctx, regToken)
	if err != nil {
		return nil, err
	}

	if err := ValidateRegistrationInput(owner, pin, currency); err != nil {
		return nil, err
	}

	pinHash, salt, err := HashPIN(pin)
	if err != nil {
		return nil, fmt.Errorf("hash PIN: %w", err)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("generate account ID: %w", err)
	}

	req := AccountCreateRequest{
		ID:          id,
		Owner:       owner,
		PhoneNumber: phone,
		PinHash:     hex.EncodeToString(pinHash),
		PinSalt:     hex.EncodeToString(salt),
		Currency:    currency,
		Balance:     0,
		CreatedAt:   time.Now(),
	}

	if err = s.Repository.CreateAccount(ctx, req); err != nil {
		return nil, err
	}

	if err := s.Repository.DeleteRegistration(ctx, regToken); err != nil {
		return nil, err
	}

	tokens, err := s.IssueTokens(ctx, id.String())
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *AuthService) GenerateJWT(id string) (string, error) {
	claims := jwt.MapClaims{
		"uid": id,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.jwtKey)
}

func (s *AuthService) IssueTokens(
	ctx context.Context,
	accountID string,
) (*TokenPair, error) {
	accessToken, err := s.GenerateJWT(accountID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	accID, err := uuid.Parse(accountID)
	if err != nil {
		return nil, err
	}

	uuid, _ := uuid.NewRandom()

	record := RefreshTokenRecord{
		ID:        uuid,
		AccountID: accID,
		TokenHash: fmt.Sprintf("%x", sha256.Sum256([]byte(refreshToken))),
		ExpiresAt: time.Now().Add(720 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.Repository.SaveToken(ctx, record); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}, nil
}
