package auth

import "errors"

var ErrTooManyAttempts = errors.New("Too many attempts")
var ErrInvalidOTP = errors.New("Invalid OTP code")
