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
	guildLimit  int = 100 // Maximum number of guilds for name change
	maxTimeouts int = 5   // Maximum number of consecutive timeouts before marking server as offline
)

// Bot configuration.
type Bot struct {
	Name           string       `yaml:"name"`            // Bot display name
	DiscordToken   string       `yaml:"discord_token"`   // Discord bot token
	UpdateInterval int          `yaml:"update_interval"` // Status update interval in seconds
	Server         types.Server `yaml:"server"`          // DayZ server configuration
}

// isStatusChanged checks if any of the server status values have changed.
func isStatusChanged(res dayz.ServerInfo, online byte, queue, time string) bool {
	return online != res.Players || queue != res.Queue || time != res.Time
}

// isDay determines if the given time string represents daytime (06:00-19:59).
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

// buildStatusString creates a formatted status string for Discord presence.
// Format: 👤 5/20 (+3) | 🌞 12:30.
func buildStateString(res *dayz.ServerInfo, emojis types.Emojis) (string, error) {
	var queueString, timeString, onlineString string

	onlineString = fmt.Sprintf(" %v %v/%v", emojis.Human, res.Players, res.MaxPlayers)

	if len(res.Queue) > 0 && res.Queue != "0" {
		queueString = fmt.Sprintf(" (+%s)", res.Queue)
	}

	if len(res.Time) > 0 {
		emoji := emojis.Night
		if ok, err := isDay(res.Time); ok {
			if err != nil {
				return "", err
			}

			emoji = emojis.Day
		}
		timeString = fmt.Sprintf(" | %v %v", emoji, res.Time)
	}

	return fmt.Sprintf("%v%v%v", onlineString, queueString, timeString), nil
}

// updateStatus updates the Discord bot's presence status.
func updateStatus(discord *discordgo.Session, status string, state string) error {
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

	return discord.UpdateStatusComplex(data)
}

// handleServerError handles server connection errors and timeout logic.
func handleServerError(err error, server *dayz.Server, maxTimeouts int, discord *discordgo.Session, log *slog.Logger, offlineText string) error {
	log.Error(
		"Failed to connect to the server. It may be turned off, or the IP address or port might be incorrect.",
		"error", err,
	)

	server.Timeout()

	if server.Timeouts >= maxTimeouts {
		err = updateStatus(discord, "dnd", offlineText)
		if err != nil {
			return err
		}
	}

	return nil
}

// Updates the bot's Discord status with actual server information.
func updateServerStatus(discord *discordgo.Session, log *slog.Logger, res *dayz.ServerInfo, emojis types.Emojis, online *byte, queue, time *string) error {
	*online, *queue, *time = res.Players, res.Queue, res.Time

	state, err := buildStateString(res, emojis)
	if err != nil {
		return err
	}

	err = updateStatus(discord, "online", state)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Bot \"%s\" has been updated", res.Name),
		"players", res.Players,
		"queue", res.Queue,
		"server-time", res.Time,
	)

	return nil
}

// Changes the bot's nickname to the configured name in all guilds.
func updateBotNameInAllGuilds(discord *discordgo.Session, log *slog.Logger, name string) error {
	guilds, err := discord.UserGuilds(guildLimit, "", "", false)
	if err != nil {
		log.Error("Failed to retrieve guilds.",
			"error", err,
		)

		return err
	}

	for _, g := range guilds {
		err = discord.GuildMemberNickname(g.ID, "@me", name)
		if err != nil {
			return err
		}
	}

	return nil
}

// Run the bot and begins monitoring the DayZ server.
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

	log.InfoContext(ctx, fmt.Sprintf("Bot for \"%s\" has started", b.Name),
		"ip", fmt.Sprintf("%v:%v", b.Server.IP, b.Server.Port),
		"interval", fmt.Sprintf("%ds", b.UpdateInterval),
	)

	err = updateBotNameInAllGuilds(discord, log, b.Name)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to update bot \"%s\" name", b.Name),
			"error", err,
		)
	}

	ticker := time.NewTicker(time.Duration(b.UpdateInterval) * time.Second)
	defer ticker.Stop()

	server := dayz.Server{
		IP:        b.Server.IP,
		QueryPort: b.Server.QueryPort,
	}

	var online byte
	var queue, time string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			res, err := server.GetServerInfo()
			if err != nil {
				err = handleServerError(err, &server, maxTimeouts, discord, log, offlineText)
				if err != nil {
					return err
				}

				continue
			}

			if server.Timeouts > 0 {
				server.ResetTimeout()
			}

			if !isStatusChanged(res, online, queue, time) {
				continue
			}

			err = updateServerStatus(discord, log, &res, emojis, &online, &queue, &time)
			if err != nil {
				return err
			}
		}
	}
}
