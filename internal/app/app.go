package app

import (
	"context"
	"log/slog"
	"playground/internal/bot"
	"playground/internal/config"
	"sync"
	"time"
)

type App struct {
	config          *config.Config
	log             *slog.Logger
	shutdownTimeout time.Duration
}

const (
	defaultShutdownTimeout = 10 * time.Second
)

func New(cfg *config.Config, log *slog.Logger) *App {
	return &App{
		config:          cfg,
		log:             log,
		shutdownTimeout: defaultShutdownTimeout,
	}
}

func (app *App) Run(ctx context.Context) error {
	app.printWelcomeMessages()

	wg := &sync.WaitGroup{}

	for _, botConfig := range app.config.Bots {
		wg.Add(1)
		go app.runBot(ctx, botConfig, wg)
	}

	<-ctx.Done()
	app.log.Info("Shutdown signal received")

	return app.gracefulShutdown(wg)
}

func (app *App) printWelcomeMessages() {
	app.log.Info("⭐ If you find this bot useful, drop a star on GitHub •`_´•")
	app.log.Info("💡 Any issues or ideas? Let me know by opening an issue.")
}

func (app *App) runBot(ctx context.Context, b bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()

	app.log.Info("Starting bot", "name", b.Name)

	err := b.Run(ctx, app.config.Emojis, app.config.OfflineText)
	if err != nil {
		if ctx.Err() != nil {
			app.log.Debug("Bot stopped due to context cancellation", "name", b.Name)
			return
		}

		app.log.Error("Bot failed", "name", b.Name, "error", err)
	}
}

func (app *App) gracefulShutdown(wg *sync.WaitGroup) error {
	app.log.Info("Initiating graceful shutdown")

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		app.log.Info("All bots shut down successfully")
		return nil
	case <-time.After(app.shutdownTimeout):
		app.log.Warn("Graceful shutdown timeout exceeded, forcing exit")
		return context.DeadlineExceeded
	}
}
