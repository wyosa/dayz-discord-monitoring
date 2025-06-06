package config

import (
	"os"
	"playground/internal/bot"
	"playground/internal/types"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Bots        []bot.Bot    `yaml:"bots"`
	Emojis      types.Emojis `yaml:"emojis"`
	OfflineText string       `yaml:"offline"`
}

func MustLoad(configPath string) (*Config, error) {
	var cfg Config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
