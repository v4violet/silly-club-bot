package poll_pin

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
)

func Init() {
	modules.RegisterModule(modules.Module{
		Name: "poll_pin",
		Init: func() []bot.ConfigOpt {
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(onMessageCreate),
				bot.WithEventListenerFunc(onMessageUpdate),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessages)),
			}
		},
	})
}

func onMessageCreate(event *events.MessageCreate) {
	if event.Message.Author.Bot || event.GuildID == nil || config.Config.Discord.GuildId != *event.GuildID {
		return
	}
	if event.Message.Poll != nil {
		if err := event.Client().Rest().PinMessage(event.ChannelID, event.MessageID); err != nil {
			slog.Error("error pinning poll message", slog.Any("error", err))
			return
		}
	}
}

func onMessageUpdate(event *events.MessageUpdate) {
	if event.Message.Author.Bot || event.GuildID == nil || config.Config.Discord.GuildId != *event.GuildID {
		return
	}
	if event.Message.Poll != nil && event.Message.Poll.Results != nil && event.Message.Poll.Results.IsFinalized {
		if err := event.Client().Rest().UnpinMessage(*event.Message.MessageReference.ChannelID, *event.Message.MessageReference.MessageID); err != nil {
			slog.Error("error unpinning poll message", slog.Any("error", err))
			return
		}
	}
}
