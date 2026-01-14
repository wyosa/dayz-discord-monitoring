package app

import (
	"context"
	"errors"
	"log/slog"
	"playground/internal/bot"
	"playground/internal/config"
	"sync"
	"time"
)

type App struct {
	config *config.Config
	log    *slog.Logger
}

const (
	timeoutTime = 10
)

func New(cfg *config.Config, log *slog.Logger) *App {
	return &App{
		config: cfg,
		log:    log,
	}
}

func (app *App) Run(ctx context.Context) error {
	app.printWelcomeMessages()

	wg := app.startBots(ctx)

	<-ctx.Done()
	app.log.Info("Shutting down all bots...")

	app.gracefulShutdown(wg)

	return nil
}

func (app *App) printWelcomeMessages() {
	app.log.Info("⭐ If you find this bot useful, drop a star on GitHub •`_´•")
	app.log.Info("💡 Any issues or ideas? Let me know by opening an issue.")
}

func (app *App) startBots(ctx context.Context) *sync.WaitGroup {
	wg := &sync.WaitGroup{}

	for _, botConfig := range app.config.Bots {
		wg.Add(1)
		go app.runBot(ctx, botConfig, wg)
	}

	return wg
}

func (app *App) runBot(ctx context.Context, b bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()

	err := b.Run(ctx, app.config.Emojis, app.config.OfflineText)
	if err != nil && !errors.Is(err, context.Canceled) {
		app.log.Error("Bot failed", "bot", b.Name, "error", err)
	}
}

func (app *App) gracefulShutdown(wg *sync.WaitGroup) {
	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		app.log.Info("All bots shut down successfully")
	case <-time.After(timeoutTime * time.Second):
		app.log.Warn("Force shutdown after timeout")
	}
}
