package main

import (
	"banking-app/internal/app"
	"banking-app/internal/config"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	app, err := app.NewApp(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}

}
