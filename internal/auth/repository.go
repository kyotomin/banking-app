package auth

import (
	"banking-app/internal/postgres"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type AuthRepository struct {
	redis *redis.Client
	db    *postgres.PostgresDatabase
}

func NewAuthRepository(client *redis.Client, db *postgres.PostgresDatabase) *AuthRepository {
	return &AuthRepository{
		redis: client,
		db:    db,
	}
}

func (r *AuthRepository) CreateChallenge(
	ctx context.Context,
	challenge *AuthChallenge,
) error {

	data, err := json.Marshal(challenge)
	if err != nil {
		return fmt.Errorf("marshal challenge: %w", err)
	}

	ttl := time.Until(challenge.ExpiresAt)
	if ttl <= 0 {
		return errors.New("challenge already expired")
	}

	key := "auth:challenge:" + challenge.ID

	if err := r.redis.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("save challenge: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetChallenge(
	ctx context.Context,
	id string,
) (*AuthChallenge, error) {
	key := "auth:challenge:" + id

	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("challenge not found")
		}

		return nil, fmt.Errorf("get challenge: %w", err)
	}

	var challenge AuthChallenge
	if err := json.Unmarshal(data, &challenge); err != nil {
		return nil, fmt.Errorf("unmarshal challenge: %w", err)
	}

	return &challenge, nil
}

func (r *AuthRepository) GetByPhone(
	ctx context.Context,
	phone string,
) (string, error) {
	var id string

	err := r.db.Pool.QueryRow(
		ctx,
		`SELECT id
		 FROM accounts
		 WHERE phone_number = $1`,
		phone,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}

		return "", fmt.Errorf("finding account by phone: %w", err)
	}

	return id, nil
}

func (r *AuthRepository) UpdateChallenge(
	ctx context.Context,
	challenge *AuthChallenge,
) error {
	data, err := json.Marshal(challenge)
	if err != nil {
		return fmt.Errorf("marshal challenge: %w", err)
	}

	ttl := time.Until(challenge.ExpiresAt)
	if ttl <= 0 {
		return errors.New("challenge already expired")
	}

	key := "auth:challenge:" + challenge.ID

	if err := r.redis.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("update challenge: %w", err)
	}

	return nil
}

func (r *AuthRepository) DeleteChallenge(
	ctx context.Context,
	id string,
) error {
	key := "auth:challenge:" + id

	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete challeng: %w", err)
	}

	return nil
}

func (r *AuthRepository) CreateRegistration(
	ctx context.Context,
	token string,
	phone string,
) error {
	key := "auth:registration:" + token

	ttl := 5 * time.Minute

	if err := r.redis.Set(ctx, key, phone, ttl).Err(); err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetRegistration(
	ctx context.Context,
	token string,
) (string, error) {
	key := "auth:registration:" + token

	phone, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", errors.New("registration not found")
		}

		return "", fmt.Errorf("finding registration by token: %w", err)
	}

	return string(phone), nil
}

func (r *AuthRepository) DeleteRegistration(
	ctx context.Context,
	token string,
) error {
	key := "auth:registration:" + token
	if err := r.redis.Del(ctx, key); err != nil {
		return fmt.Errorf("registration not exist or already deleted: %w", err)
	}

	return nil
}

func (r *AuthRepository) CreateAccount(
	ctx context.Context,
	req AccountCreateRequest,
) error {
	query := `
		INSERT INTO accounts
		(id, owner, phone_number, secret_code_hash, secret_code_salt, currency, balance)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		req.ID,
		req.Owner,
		req.PhoneNumber,
		req.PinHash,
		req.PinSalt,
		req.Currency,
		req.Balance,
		req.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("account creation request: %w", err)
	}

	return nil
}

func (r *AuthRepository) SaveToken(
	ctx context.Context,
	token RefreshTokenRecord,
) error {
	query := `
		INSERT INTO refresh_tokens
		(id, account_id, token_hash, expires_at, created_at)
		 VALUES($1, $2, $3, $4, $5;
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		token.ID,
		token.AccountID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}

	return nil
}
