package postgres

import (
	"banking-app/internal/config"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDatabase struct {
	Pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, cfg config.Config) (*PostgresDatabase, error) {
	pool, err := pgxpool.New(ctx, cfg.PG.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to start database.")
	}

	if err := pingWithRetry(ctx, pool, cfg.PG.MaxRetryAttempts); err != nil {
		return nil, err
	}

	return &PostgresDatabase{
		Pool: pool,
	}, nil
}

func pingWithRetry(ctx context.Context, pool *pgxpool.Pool, maxAttempts int) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := pool.Ping(ctx); err == nil {
			slog.Debug("Conntected to postgres")
			return nil
		} else if attempt == maxAttempts {
			return fmt.Errorf("postgres is unavailable: %w", err)
		}

		timer := time.NewTimer(1 * time.Second)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}

	return nil
}
