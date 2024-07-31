//go:build modules.all || modules.vote_pin

package modules

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/sethvargo/go-envconfig"
	"github.com/v4violet/silly-club-bot/config"
)

var VotePinConfig struct {
	Emoji    string `env:"MODULES_VOTE_PIN_EMOJI,default=📌"`
	MinVotes int    `env:"MODULES_VOTE_PIN_MIN_VOTES,default=5"`
}

func init() {
	Modules["vote_pin"] = Module{
		Init: func() ([]bot.ConfigOpt, error) {
			if err := envconfig.Process(context.Background(), &VotePinConfig); err != nil {
				return nil, err
			}

			return []bot.ConfigOpt{
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessageReactions)),
				bot.WithEventListenerFunc(func(event *events.GuildMessageReactionAdd) {

					if event.Emoji.Name == nil || event.Emoji.Reaction() != VotePinConfig.Emoji || event.GuildID != config.Config.Discord.GuildId {
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
						if event.Emoji.Reaction() == VotePinConfig.Emoji {
							count = reaction.Count
							break
						}
					}
					if count >= VotePinConfig.MinVotes {
						event.Client().Rest().PinMessage(event.ChannelID, event.MessageID)
					}
				}),
			}, nil
		},
	}
}
