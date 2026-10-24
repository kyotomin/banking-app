package transfers

import (
	"banking-app/internal/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type TransfersRepository struct {
	db *postgres.PostgresDatabase
}

func NewTransfersRepository(db *postgres.PostgresDatabase) *TransfersRepository {
	return &TransfersRepository{
		db: db,
	}
}

func (r *TransfersRepository) CreateTransfer(
	ctx context.Context,
	fromAccountID string,
	toAccountID string,
	amount string,
	idempotencyKey string,
) (string, error) {
	if fromAccountID == toAccountID {
		return "", errors.New("cannot transfer to same account")
	}

	amountDecimal, err := decimal.NewFromString(amount)
	if err != nil || !amountDecimal.IsPositive() {
		return "", errors.New("invalid transfer amount")
	}

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`
		SELECT (id, balance, currency)
		FROM accounts
		WHERE id IN ($1, $2)
		ORDER BY id
		FOR UPDATE
		`,
		fromAccountID,
		toAccountID,
	)
	if err != nil {
		return "", fmt.Errorf("locking accounts: %w", err)
	}

	fromInfo, fromErr := r.GetAccountInfo(ctx, tx, fromAccountID)
	toInfo, toErr := r.GetAccountInfo(ctx, tx, toAccountID)

	if fromErr != nil || toErr != nil {
		return "", ErrAccountNotFound
	}

	if fromInfo.Currency != toInfo.Currency {
		return "", ErrCurrencyMismatch
	}

	balance, err := decimal.NewFromString(fromInfo.Balance)
	if err != nil {
		return "", ErrParseFailed
	}

	amountDecimal, err = decimal.NewFromString(amount)
	if err != nil {
		return "", ErrParseFailed
	}

	if balance.LessThan(amountDecimal) {
		return "", ErrInsufficientFunds
	}

	var transferID string

	err = tx.QueryRow(ctx, `
		INSERT INTO transfers (
		from_account_id,
		to_account_id,
		amount,
		idempotency_key
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING id
	`,
		fromAccountID,
		toAccountID,
		amount,
		idempotencyKey,
	).Scan(&transferID)

	if errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)

		err := r.db.Pool.QueryRow(ctx, `
			SELECT id
			FROM transfers
			WHERE idempotency_key = $1
		`,
			idempotencyKey,
		).Scan(&transferID)

		if err != nil {
			return "", fmt.Errorf("get existing transfer: %w", err)
		}

		return transferID, nil
	}

	if err != nil {
		return "", fmt.Errorf("create transfer: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE accounts
		SET balance = balance - $1
		WHERE id = $2
	`,
		amountDecimal.String(),
		fromAccountID,
	)
	if err != nil {
		return "", fmt.Errorf("debit sender: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
	`,
		amountDecimal.String(),
		toAccountID,
	)

	_, err = tx.Exec(ctx, `
		INSERT INTO entries (transfer_id, account_id, amount, type)
		VALUES ($1, $2, -$3, 'transfer_out')
	`, transferID, fromAccountID, amountDecimal.String())
	if err != nil {
		return "", fmt.Errorf("create outgoing entry: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO entries (transfer_id, account_id, amount, type)
		VALUES ($1, $2, -$3, 'transfer_in')
	`, transferID, toAccountID, amountDecimal.String())
	if err != nil {
		return "", fmt.Errorf("create outgoing entry: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit transfer: %w", err)
	}

	return transferID, nil
}

func (r *TransfersRepository) GetAccountInfo(
	ctx context.Context,
	tx pgx.Tx,
	accountID string,
) (AccountInfo, error) {
	var account AccountInfo

	err := tx.QueryRow(ctx,
		`SELECT balance, currency
         FROM accounts
         WHERE id = $1
		 FOR UPDATE`,
		accountID,
	).Scan(&account.Balance, &account.Currency)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AccountInfo{}, ErrAccountNotFound
		}
		return AccountInfo{}, fmt.Errorf("get account info: %w", err)
	}

	return account, nil
}

func (r *TransfersRepository) GetByIdempotencyKey(
	ctx context.Context,
	idempotencyKey string,
) (*Transfer, error) {
	var info Transfer

	err := r.db.Pool.QueryRow(ctx, `
		SELECT (id, from_account_id, to_account_id, amount)
		FROM transfers
		WHERE idempotency_key = $1
	`, idempotencyKey).Scan(&info)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("get transfer by idempotency key: %w", err)
	}

	return &info, nil
}
