package app

import (
	"banking-app/internal/auth"
	"banking-app/internal/config"
	"banking-app/internal/dashboard"
	"banking-app/internal/postgres"
	"banking-app/internal/redis"
	"banking-app/internal/telegram"
	"banking-app/internal/transfers"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	router *gin.Engine
	close  func()
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	appCtx := context.Background()

	postgresDB, err := postgres.NewPostgres(appCtx, *cfg)
	if err != nil {
		return nil, err
	}

	redisDB, err := redis.New(appCtx, cfg.RE.Addr, cfg.RE.Password)
	if err != nil {
		postgresDB.Pool.Close()
		return nil, err
	}

	// Auth Register
	authRepository := auth.NewAuthRepository(redisDB, postgresDB)

	TelegramGatewaySender := telegram.NewGateway(cfg.Secret.TGToken)
	authService := auth.NewAuthService(authRepository, cfg.Secret.OTPSecret, cfg.Secret.JWTKey, TelegramGatewaySender)

	authHandler := auth.NewAuthHandler(authService)

	transfersRepository := transfers.NewTransfersRepository(postgresDB)
	transfersService := transfers.NewTransfersService(transfersRepository)
	transfersHandler := transfers.NewTransfersHandler(transfersService)

	dashboardRepository := dashboard.NewDashboardRepository(postgresDB, redisDB)
	dashboardService := dashboard.NewDashboardService(dashboardRepository)
	dashboardHandler := dashboard.NewDashboardHandler(dashboardService)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	authHandler.RegisterRoutes(router)
	transfersHandler.RegisterRoutes(router)
	dashboardHandler.RegisterRoutes(router)

	return &App{
		router: router,
		close: func() {
			redisDB.Close()
			postgresDB.Pool.Close()
		},
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":8080",
		Handler: a.router,
	}

	errCh := make(chan error, 1)

	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			server.Close()
			return err
		}

		err = <-errCh
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}
}

func (a *App) Close() {
	if a.close != nil {
		a.close()
	}
}
