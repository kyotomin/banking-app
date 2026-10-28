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

func (r *DashboardRepository) GetMonthlyStats(
	ctx context.Context,
	accountID string,
) (*MonthlyStats, error) {
	var stats MonthlyStats

	err := r.db.Pool.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(amount) FILTER(WHERE type = 'deposit'), 0),
			COALESCE(SUM(amount) FILTER(WHERE type = 'withdrawal'), 0)
		FROM entries
		WHERE account_id = $1
		AND created_at >= date_trunc('month', now())
		AND created < date_trunc('month', now()) + interval '1 month'
	`, accountID).Scan(
		&stats.TotalIncome,
		&stats.TotalSpent,
	)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, transfer_id, amount, type, category, created_at
		FROM entries
		WHERE account_id = %1
		AND created_at >= date_trunc('month', now())
		AND created_at < date_trunc('month', now()) + interval '1 month'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry Entry

		if err := rows.Scan(
			&entry.ID,
			&entry.Amount,
			&entry.Type,
			&entry.Category,
			&entry.CreatedAt,
		); err != nil {
			return nil, err
		}

		stats.Entries = append(stats.Entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *DashboardRepository) GetCategoryStats(
	ctx context.Context,
	accountID string,
) (*CategoryStats, error) {
	var stats CategoryStats

	rows, err := r.db.Pool.Query(ctx, `
		SELECT 
			category,
			COALESCE(SUM(ABS(amount)) AS total_spent
		FROM entries
		WHERE account_id = $1
			AND type = 'payment'
			AND created_at >= date_trunc('month', now())
			AND created_at < date_trunc('month', now()) + interval '1 month'
		GROUP BY category
		ORDER BY total_spent DESC
	`, accountID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var category Category

		if err := rows.Scan(
			&category.Title,
			&category.TotalSpent,
		); err != nil {
			return nil, err
		}

		stats.Categories = append(stats.Categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &stats, nil
}
