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
	Challenge *ChallengeRepository
	otpSecret string
	Account   *AccountRepository
}

const (
	otpTTL         = 5 * time.Minute
	maxOTPAttempts = 5
)

func (s *AuthService) StartAuth(
	ctx context.Context,
	rawPhone string,
) (*AuthChallenge, error) {

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

	if err := s.Challenge.Create(ctx, challenge); err != nil {
		return nil, err
	}

	//TODO: Add OTP send

	return challenge, nil
}

func (s *AuthService) VerifyCode(
	ctx context.Context,
	challengeID string,
	code string,
) error {
	challenge, err := s.Challenge.Get(ctx, challengeID)
	if err != nil {
		return err
	}

	if challenge.Attempts >= maxOTPAttempts {
		if err := s.Challenge.Delete(ctx, challengeID); err != nil {
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
			if err := s.Challenge.Delete(ctx, challengeID); err != nil {
				return fmt.Errorf("delete exhausted challenge: %w", err)
			}
			return ErrTooManyAttempts
		}

		if err := s.Challenge.Update(ctx, challenge); err != nil {
			return fmt.Errorf("update challenge attempts: %w", err)
		}

		return ErrInvalidOTP
	}

	if err := s.Challenge.Delete(ctx, challengeID); err != nil {
		return fmt.Errorf("consume challenge: %w", err)
	}

	return nil
}
