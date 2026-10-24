package transfers

import (
	"context"

	"github.com/shopspring/decimal"
)

type TransfersService struct {
	repository *TransfersRepository
}

func NewTransfersService(repo *TransfersRepository) *TransfersService {
	return &TransfersService{
		repository: repo,
	}
}

func (s *TransfersService) IssueTransfer(
	ctx context.Context,
	info TransferInfo,
) (string, error) {
	if info.FromAccount == info.ToAccount {
		return "", ErrInvalidRecipient
	}

	amount, err := decimal.NewFromString(info.Amount)
	if err != nil {
		return "", ErrInvalidAmount
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return "", ErrInvalidAmount
	}

	transfer, err := s.repository.GetByIdempotencyKey(ctx, info.IKey)
	if err != nil {
		return "", err
	}

	if transfer != nil {
		existing_amount, err := decimal.NewFromString(transfer.Amount)
		if err != nil {
			return "", err
		}

		if !amount.Equal(existing_amount) || transfer.FromAccount != info.FromAccount ||
			transfer.ToAccount != info.ToAccount {
			return "", ErrIdempotencyConflict
		}

		return transfer.ID, nil
	}

	return s.repository.CreateTransfer(
		ctx,
		info.FromAccount,
		info.ToAccount,
		info.Amount,
		info.IKey,
	)
}
