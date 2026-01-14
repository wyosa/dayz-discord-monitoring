package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"playground/internal/app"
	"playground/internal/config"
	"playground/internal/prettylog"
)

func main() {
	log := prettylog.NewLogger(slog.LevelDebug, false)

	path, err := config.Parse()
	if err != nil {
		log.Error("Config parsing failed", "error", err)
		os.Exit(1)
	}

	cfg := config.MustLoad(path)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	application := app.New(cfg, log)

	if err = application.Run(ctx); err != nil {
		log.Error("Application failed", "error", err)
		defer os.Exit(1)
	}
}
