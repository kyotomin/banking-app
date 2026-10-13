package auth

import (
	"context"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	Repository *AuthRepository

	otpSecret string
	sender    OTPSender
}

func NewAuthService(repo *AuthRepository, secret string, sender OTPSender) *AuthService {
	return &AuthService{
		Repository: repo,
		otpSecret:  secret,
		sender:     sender,
	}
}

const (
	otpTTL         = 5 * time.Minute
	maxOTPAttempts = 5
)

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
		CodeHash:    string(HashOTP(s.otpSecret, id, code)),
		ExpiresAt:   now.Add(otpTTL),
		Attempts:    5,
		CreatedAt:   now,
	}

	if err := s.Repository.Create(ctx, challenge); err != nil {
		return nil, err
	}

	if err := s.sender.SendCode(ctx, phone, code); err != nil {
		if delErr := s.Repository.Delete(ctx, id); delErr != nil {
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
) error {
	challenge, err := s.Repository.Get(ctx, challengeID)
	if err != nil {
		return err
	}

	if challenge.Attempts >= maxOTPAttempts {
		if err := s.Repository.Delete(ctx, challengeID); err != nil {
			return fmt.Errorf("delete exhausted challenge: %w", err)
		}
		return ErrTooManyAttempts
	}

	expected := HashOTP(s.otpSecret, challengeID, code)
	actual, err := hex.DecodeString(challenge.CodeHash)
	if err != nil {
		return err
	}

	if !hmac.Equal(expected, actual) {
		challenge.Attempts++

		if challenge.Attempts >= maxOTPAttempts {
			if err := s.Repository.Delete(ctx, challengeID); err != nil {
				return fmt.Errorf("delete exhausted challenge: %w", err)
			}
			return ErrTooManyAttempts
		}

		if err := s.Repository.Update(ctx, challenge); err != nil {
			return fmt.Errorf("update challenge attempts: %w", err)
		}

		return ErrInvalidOTP
	}

	if err := s.Repository.Delete(ctx, challengeID); err != nil {
		return fmt.Errorf("consume challenge: %w", err)
	}

	return nil
}
