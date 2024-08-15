//go:build modules.all || modules.vote_pin

package modules

import (
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"go.uber.org/fx"
)

type VotePinConfig struct {
	Emoji    string `env:"MODULES_VOTE_PIN_EMOJI" envDefault:"📌"`
	MinVotes int    `env:"MODULES_VOTE_PIN_MIN_VOTES" envDefault:"5"`
}

func init() {
	modules = append(modules, fx.Module("modules/vote_pin",
		fx.Provide(
			env.ParseAs[VotePinConfig],
			ProvideVotePin,
		),
		fx.Invoke(NewVotePin),
	))
}

func ProvideVotePin() Results {
	return Results{
		Options: []bot.ConfigOpt{
			bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessageReactions)),
		},
	}
}

func NewVotePin(p ParamsWithConfig[VotePinConfig]) {
	p.Client.AddEventListeners(bot.NewListenerFunc(func(event *events.GuildMessageReactionAdd) {

		if event.Emoji.Name == nil || event.Emoji.Reaction() != p.Config.Emoji || event.GuildID != p.DiscordConfig.GuildId {
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
			if event.Emoji.Reaction() == p.Config.Emoji {
				count = reaction.Count
				break
			}
		}
		if count >= p.Config.MinVotes {
			event.Client().Rest().PinMessage(event.ChannelID, event.MessageID)
		}
	}))
}
