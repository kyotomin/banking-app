package auth

import "context"

type OTPSender interface {
	SendCode(ctx context.Context, phone, code string) error
}
