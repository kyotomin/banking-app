package transfers

import "errors"

var (
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInvalidRecipient    = errors.New("invalid recipent. can't transfer to oneself")
	ErrParseFailed         = errors.New("can't parse current value")
	ErrAccountNotFound     = errors.New("account not found")
	ErrIdempotencyConflict = errors.New("idempotency key conflict occured")
)
