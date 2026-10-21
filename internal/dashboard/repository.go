package dashboard

import (
	"banking-app/internal/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type DashboardRepository struct {
	db    *postgres.PostgresDatabase
	redis *redis.Client
}

func NewDashboardRepository(db *postgres.PostgresDatabase, redis *redis.Client) *DashboardRepository {
	return &DashboardRepository{
		db:    db,
		redis: redis,
	}
}

func (r *DashboardRepository) GetAccount(
	ctx context.Context,
	accountID string,
) (Account, error) {
	var acc Account

	err := r.db.Pool.QueryRow(
		ctx,
		`
		SELECT (owner, balance, currency)
		FROM accounts
		WHERE id = $1
		`,
		accountID,
	).Scan(&acc)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, fmt.Errorf("account not found")
		}
		return Account{}, fmt.Errorf("accoutn not found")
	}

	return acc, nil
}

func (r *DashboardRepository) GetTotal(
	ctx context.Context,
	accountID string,
) (TotalSpent, error) {
	var total TotalSpent

	err := r.db.Pool.QueryRow(
		ctx,
		`SELECT COALESCE(SUM(ABS(amount)), 0)
		 FROM entries
		 WHERE account_id = $1
		 AND type IN ('withdrawal', 'transfer_out')
		 AND created_at >= date_trunc('month', now())
		 AND created_at < date_trunc('month', now()) + interval '1 month' 
		`,
		accountID,
	).Scan(&total.Amount)

	if err != nil {
		return TotalSpent{}, ErrAccountNotFound
	}

	return total, nil
}
