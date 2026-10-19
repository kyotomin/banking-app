package auth

import "errors"

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrInvalidPIN           = errors.New("invalid PIN")
	ErrInvalidCredentials   = errors.New("invalid phone or PIN")
	ErrInvalidOTP           = errors.New("invalid OTP")
	ErrTooManyAttempts      = errors.New("too many attempts")
	ErrChallengeNotFound    = errors.New("challenge not found")
	ErrRegistrationNotFound = errors.New("registration not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrTokenExpired         = errors.New("token expired")
	ErrInvalidToken         = errors.New("invalid token")
	ErrPhoneAlreadyExists   = errors.New("phone already exists")
)
