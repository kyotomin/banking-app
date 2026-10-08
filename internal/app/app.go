package app

import "banking-app/internal/config"

type App struct {
	Config config.Config
}

func New(cfg config.Config) *App {
	return &App{
		Config: cfg,
	}
}

func (a *App) Start() error {
	return nil
}
