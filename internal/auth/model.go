package auth

import (
	"time"
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
