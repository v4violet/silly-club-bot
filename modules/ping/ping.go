package ping

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
)

func Init() {
	modules.RegisterModule(modules.Module{
		Name: "ping",
		Init: func() []bot.ConfigOpt {
			if config.Config.Discord.AuthorId == nil {
				slog.Warn("`ping` module enabled but missing author id (skipping init)")
				return []bot.ConfigOpt{}
			}
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(onMessageCreate),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessages, gateway.IntentMessageContent)),
			}
		},
	})
}

func onMessageCreate(event *events.MessageCreate) {
	if event.Message.Author.Bot || event.GuildID == nil || config.Config.Discord.GuildId != *event.GuildID {
		return
	}
	if config.Config.Discord.AuthorId == nil || event.Message.Author.ID != *config.Config.Discord.AuthorId {
		return
	}
	if event.Message.Content != "!ping" {
		return
	}
	if _, err := event.Client().Rest().CreateMessage(event.ChannelID, discord.MessageCreate{
		Content: fmt.Sprintf("pong! \ngateway: %v", event.Client().Gateway().Latency()),
		MessageReference: &discord.MessageReference{
			MessageID:       &event.MessageID,
			ChannelID:       &event.ChannelID,
			GuildID:         event.GuildID,
			FailIfNotExists: false,
		},
	}); err != nil {
		slog.Error("error sending pong",
			slog.Any("error", err),
			slog.Any("channel_id", event.ChannelID),
			slog.Any("message_id", event.MessageID),
		)
	}
}
