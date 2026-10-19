package auth

import (
	"time"

	"github.com/google/uuid"
)

type StartAuthRequest struct {
	PhoneNumber string `json:"phone_number"`
}

type AuthChallenge struct {
	ID          string
	PhoneNumber string
	CodeHash    string
	ExpiresAt   time.Time
	Attempts    int
	CreatedAt   time.Time
}

type SafeAuthChallenge struct {
	ID        string
	ExpiresAt time.Time
}

type VerifyCodeRequest struct {
	ChallengeID string `json:"challenge_id" binding:"required"`
	Code        string `json:"code" binding:"required"`
}

type AccountExistResponse struct {
	ID string
}

type VerificationResult struct {
	Status string `json:"status"`
	Token  string `json:"registration_token,omitempty"`
}

type AccountCreateRequest struct {
	ID          uuid.UUID
	Owner       string
	PhoneNumber string
	Balance     float64
	PinHash     string
	PinSalt     string
	Currency    string
	CreatedAt   time.Time
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type RefreshTokenRecord struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type RegisterAccountRequest struct {
	Token    string `json:"token"`
	Owner    string `json:"owner"`
	Pin      string `json:"pin"`
	Currency string `json:"currency"`
}

type LoginAccountRequest struct {
	PhoneNumber string `json:"phone_number"`
	Pin         string `json:"pin"`
}

type AccountCredentials struct {
	ID      string
	PinHash string
	PinSalt string
}
