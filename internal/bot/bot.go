package bot

import (
	"context"
	"fmt"
	"log/slog"
	"playground/internal/dayz"
	"playground/internal/prettylog"
	"playground/internal/types"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	GUILD_LIMIT = 100
)

func updateStatus(s *discordgo.Session, status string, state string) (err error) {
	data := discordgo.UpdateStatusData{
		Status: status,
	}

	if state != "" {
		data.Activities = []*discordgo.Activity{{
			Name:  "0_o",
			Type:  discordgo.ActivityTypeCustom,
			State: state,
		}}
	}

	return s.UpdateStatusComplex(data)
}

type Bot struct {
	Name           string       `yaml:"name"`
	DiscordToken   string       `yaml:"discord_token"`
	UpdateInterval int          `yaml:"update_interval"`
	Server         types.Server `yaml:"server"`
}

func (b *Bot) Run(ctx context.Context, emojis types.Emojis, offlineText string) error {
	log := prettylog.NewLogger(slog.LevelDebug, false)

	discord, err := discordgo.New("Bot " + b.DiscordToken)
	if err != nil {
		return err
	}

	err = discord.Open()
	if err != nil {
		return err
	}
	defer discord.Close()

	log.Info(fmt.Sprintf("Bot for \"%s\" started successfully", b.Name),
		"name", b.Name,
		"ip", fmt.Sprintf("%v:%v", b.Server.Ip, b.Server.Port),
		"interval", fmt.Sprintf("%ds", b.UpdateInterval),
	)

	// получение всех серверов, на которых находится бот.
	guilds, err := discord.UserGuilds(GUILD_LIMIT, "", "", false)
	if err != nil {
		log.Error("Do not disturb.",
			"error", err,
		)
	}

	// изменение имени боту (на имя из конфига) на всех серверах, где он находится.
	for _, g := range guilds {
		err := discord.GuildMemberNickname(g.ID, "@me", b.Name)
		if err != nil {
			return err
		}
	}

	ticker := time.NewTicker(time.Duration(b.UpdateInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			res, err := dayz.GetServerInfo(b.Server.Ip, b.Server.QueryPort)
			if err != nil {
				log.Error(
					"Server is offline, or the IP address or port is incorrect.",
					"error", err,
				)

				updateStatus(
					discord,
					"dnd",
					offlineText,
				)

				break
			}

			updateStatus(
				discord,
				"online",
				fmt.Sprintf("%s %v/%v", emojis.Human, res.Server.Players, res.Server.MaxPlayers),
			)
		}
	}
}
