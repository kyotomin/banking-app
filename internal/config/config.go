package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env" env:"APP_ENV" env-default:"development"`

	PG PGConfig
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

func Get() (*Config, error) {
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
