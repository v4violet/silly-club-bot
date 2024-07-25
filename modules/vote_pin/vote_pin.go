package vote_pin

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
)

const pinEmoji = "📌"
const minVotes = 5

func Init() {
	modules.RegisterModule(modules.Module{
		Name: "vote_pin",
		Init: func() []bot.ConfigOpt {
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(onMessageReactionAdd),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessageReactions)),
			}
		},
	})
}

func onMessageReactionAdd(event *events.GuildMessageReactionAdd) {
	if event.Emoji.Name == nil || *event.Emoji.Name != pinEmoji || event.GuildID != config.Config.Discord.GuildId {
		return
	}
	message, err := event.Client().Rest().GetMessage(event.ChannelID, event.MessageID)
	if err != nil {
		slog.Error("error getting reaction message", slog.Any("message_id", event.MessageID))
		return
	}
	if message.Pinned {
		return
	}
	count := 0
	for _, reaction := range message.Reactions {
		if reaction.Emoji.Name == pinEmoji {
			count = reaction.Count
			break
		}
	}
	if count >= minVotes {
		event.Client().Rest().PinMessage(event.ChannelID, event.MessageID)
	}
}
