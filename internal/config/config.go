package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env       string `yaml:"env" env:"APP_ENV" env-default:"development"`
	HMACKey   string `yaml:"HMAC_KEY" env:"HMAC_KEY"`
	OTPSecret string `yaml:"OTP_SECRET" env:"OTP_SECRET"`
	TGToken   string `yaml:"TELEGRAM_TOKEN" env:"TELEGRAM_TOKEN"`

	PG PGConfig
	RE RedisConfig
}

type PGConfig struct {
	User             string `yaml:"user"     env:"POSTGRES_USER"     env-required:"true"`
	Password         string `yaml:"password" env:"POSTGRES_PASSWORD" env-required:"true"`
	Host             string `yaml:"host"     env:"POSTGRES_HOST"     env-default:"localhost"`
	Port             string `yaml:"port"     env:"POSTGRES_PORT"     env-default:"5432"`
	Name             string `yaml:"name"     env:"POSTGRES_DB"       env-required:"true"`
	SSLMode          string `yaml:"sslmode"  env:"POSTGRES_SSLMODE"  env-default:"disable"`
	MaxRetryAttempts int    `yaml:"retries" env:"MAX_RETRY_ATTEMPTS" env-default:"5"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr" env:"REDIS_HOST" env-required:"true"`
	Password string `yaml:"password" env:"REDIS_PASSWORD" env-required:"true"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig("config.yaml", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("failed to load config. probably config is not specified.")
		}
	}

	return &cfg, nil
}

func (pg *PGConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		pg.User, pg.Password, pg.Host, pg.Port, pg.Name, pg.SSLMode,
	)
}

func (pg *PGConfig) MigrateDSN() string {
	return fmt.Sprintf(
		"pgx5://%s:%s@%s:%s/%s?sslmode=%s",
		pg.User, pg.Password, pg.Host, pg.Port, pg.Name, pg.SSLMode,
	)
}
