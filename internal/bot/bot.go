package bot

import (
	"context"
	"fmt"
	"log/slog"
	"playground/internal/dayz"
	"playground/internal/prettylog"
	"playground/internal/types"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	GUILD_LIMIT = 100
)

type Bot struct {
	Name           string       `yaml:"name"`
	DiscordToken   string       `yaml:"discord_token"`
	UpdateInterval int          `yaml:"update_interval"`
	Server         types.Server `yaml:"server"`
}

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

func isDay(time string) (bool, error) {
	res := strings.Split(time, ":")

	hour, err := strconv.Atoi(res[0])
	if err != nil {
		return false, err
	}

	if hour >= 6 && hour < 20 {
		return true, nil
	}

	return false, nil
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

	getServerInfo := dayz.GetServerInfo(b.Server.Ip, b.Server.QueryPort)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			timeouts, res, err := getServerInfo()
			if err != nil {
				log.Error(
					"Server is offline, or the IP address or port is incorrect.",
					"error", err,
				)

				if timeouts >= 5 {
					updateStatus(
						discord,
						"dnd",
						offlineText,
					)
				}

				continue
			}

			online := fmt.Sprintf(" %v %v/%v", emojis.Human, res.Players, res.MaxPlayers)
			queue := ""
			time := ""

			if res.Queue != "" && res.Queue != "0" {
				queue = fmt.Sprintf(" (+%s)", res.Queue)
			}

			if res.Time != "" {
				if ok, _ := isDay(res.Time); ok {
					time = fmt.Sprintf(" | %v %v", emojis.Day, res.Time)
				} else {
					time = fmt.Sprintf(" | %v %v", emojis.Night, res.Time)
				}
			}

			updateStatus(
				discord,
				"online",
				fmt.Sprintf("%v%v%v", online, queue, time),
			)
		}
	}
}
