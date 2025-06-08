package config

import (
	"os"
	"playground/internal/bot"
	"playground/internal/types"

	"github.com/ilyakaznacheev/cleanenv"
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
		panic(err)
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}
