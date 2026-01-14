package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"playground/internal/bot"
	"playground/internal/types"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	minUpdateInterval = 5
)

// Config represents the main application configuration structure.
// loaded from YAML configuration file.
type Config struct {
	Bots        []bot.Bot    `yaml:"bots"`    // List of Discord bots to run
	Emojis      types.Emojis `yaml:"emojis"`  // Emoji configuration for state display
	OfflineText string       `yaml:"offline"` // Text to display when server is offline
}

func MustLoad(configPath string) *Config {
	var cfg Config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Errorf("config file not found: %s", configPath))
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Errorf("failed to parse config: %w", err))
	}

	if err := cfg.Validate(); err != nil {
		panic(fmt.Errorf("invalid config: %w", err))
	}

	return &cfg
}

func Parse() (string, error) {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to config file (YAML)")

	flag.Parse()

	if configPath == "" {
		return "", errors.New("\"-config\" flag was not specified.")
	}

	return configPath, nil
}

func (c *Config) Validate() error {
	if len(c.Bots) == 0 {
		return errors.New("no bots configured")
	}

	for i, bot := range c.Bots {
		if err := validateBot(&bot, i); err != nil {
			return err
		}
	}

	if len(strings.TrimSpace(c.OfflineText)) == 0 {
		return errors.New("offline text cannot be empty")
	}

	return nil
}

func validateBot(b *bot.Bot, index int) error {
	if strings.TrimSpace(b.Name) == "" {
		return fmt.Errorf("bot[%d]: name cannot be empty", index)
	}

	if strings.TrimSpace(b.DiscordToken) == "" {
		return fmt.Errorf("bot[%d]: discord token cannot be empty", index)
	}

	if b.UpdateInterval <= 0 {
		return fmt.Errorf("bot[%d]: update interval must be positive", index)
	}

	if b.UpdateInterval < minUpdateInterval {
		return fmt.Errorf("bot[%d]: update interval must be at least 5 seconds to avoid rate limiting", index)
	}

	return nil
}
