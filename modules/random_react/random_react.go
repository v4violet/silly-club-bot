package random_react

import (
	"log/slog"
	"math/rand/v2"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
)

const percentage = 0.01
const emoji = "trollos:1144404046288007260"

func Init() {
	modules.RegisterModule(modules.Module{
		Name: "random_react",
		Init: func() []bot.ConfigOpt {
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(onMessageCreate),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessages)),
			}
		},
	})
}

func onMessageCreate(event *events.MessageCreate) {
	if event.Message.Author.Bot || event.GuildID == nil || config.Config.Discord.GuildId != *event.GuildID {
		return
	}

	_, channelWhitelisted := config.Config.Modules.RandomReact.WhitelistedChannelIds[event.ChannelID]

	if rand.Float32() <= (percentage/100) && channelWhitelisted {
		go func() {
			if err := event.Client().Rest().AddReaction(event.ChannelID, event.MessageID, emoji); err != nil {
				slog.Error("error adding reaction",
					slog.Any("error", err),
					slog.Any("channel_id", event.ChannelID),
					slog.Any("message_id", event.MessageID),
				)
				return
			}
			if _, err := event.Client().Rest().CreateMessage(*config.Config.Modules.RandomReact.LogChannelId, discord.MessageCreate{
				Content: event.Message.JumpURL(),
			}); err != nil {
				slog.Error("error logging random reaction",
					slog.Any("error", err),
					slog.Any("channel_id", event.ChannelID),
					slog.Any("message_id", event.MessageID),
				)
			}
		}()
	}
}
