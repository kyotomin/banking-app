package auth

import (
	"banking-app/internal/postgres"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

func (r *AuthRepository) Create(
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

func (r *AuthRepository) Get(
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

func (r *AuthRepository) Update(
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

func (r *AuthRepository) Delete(
	ctx context.Context,
	id string,
) error {
	key := "auth:challenge:" + id

	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete challeng: %w", err)
	}

	return nil
}
