package redis

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
)

func New(ctx context.Context, addr, password string) (*redislib.Client, error) {
	client := redislib.NewClient(&redislib.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}
