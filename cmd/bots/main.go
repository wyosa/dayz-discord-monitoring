package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"playground/internal/bot"
	"playground/internal/config"
	"playground/internal/prettylog"
	"sync"
	"time"
)

var (
	configPath = flag.String("config", "", "Path to config file (YAML)")
)

func main() {
	flag.Parse()

	log := prettylog.NewLogger(slog.LevelDebug, false)

	if *configPath == "" {
		log.Error("\"-config\" flag was not specified.")
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := config.MustLoad(*configPath)

	wg := &sync.WaitGroup{}

	for _, b := range cfg.Bots {
		monitoringBot := bot.Bot{
			Name:           b.Name,
			DiscordToken:   b.DiscordToken,
			UpdateInterval: b.UpdateInterval,
			Server:         b.Server,
		}

		wg.Add(1)
		go func(b bot.Bot) {
			defer wg.Done()
			err := b.Run(ctx, cfg.Emojis, cfg.OfflineText)
			if err != nil && err != ctx.Err() {
				log.Error("Bot failed", "bot", b.Name, "error", err)
			}
		}(monitoringBot)
	}

	<-ctx.Done()
	log.Info("Shutting down all bots...")

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info("All bots shut down successfully")
	case <-time.After(10 * time.Second):
		log.Warn("Force shutdown after timeout.")
	}
}
