//go:build modules.all || modules.poll_pin

package modules

import (
	"log/slog"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"go.uber.org/fx"
)

type PollPinConfig struct {
	// 7 days (168 hours) + 1 hour
	MaxPollDuration time.Duration `env:"MODULES_POLL_PIN_MAX_POLL_DURATION" envDefault:"169h"`
}

func init() {
	modules = append(modules, fx.Module("modules/poll_pin",
		fx.Provide(
			env.ParseAs[PollPinConfig],
			ProvidePollPin,
		),
		fx.Invoke(NewPollPin),
	))
}

func ProvidePollPin() Results {
	return Results{
		Options: []bot.ConfigOpt{
			bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessages)),
		},
	}
}

const SEVEN_DAYS_ONE_HOUR = ((7 * 24 * time.Hour) + time.Hour)

func NewPollPin(p ParamsWithConfig[PollPinConfig]) {
	p.Client.AddEventListeners(
		bot.NewListenerFunc(func(event *events.GuildMessageCreate) {
			if event.Message.Author.Bot || event.GuildID != p.DiscordConfig.GuildId || event.Message.Poll == nil {
				return
			}
			if event.Message.Poll.Expiry.Sub(time.Now()) > p.Config.MaxPollDuration {
				event.Client().Rest().AddReaction(event.ChannelID, event.MessageID, "⌛")
				return
			}
			pins, err := event.Client().Rest().GetPinnedMessages(event.ChannelID)
			if err != nil {
				slog.Error("error listing channel pins")
				return
			}
			poll_pins := []discord.Message{}
			for _, pin := range pins {
				if pin.Poll != nil {
					poll_pins = append(poll_pins, pin)
				}
			}
			if len(poll_pins) > 5 {
				slog.Warn("too many pinned polls", slog.Any("channel_id", event.ChannelID))
				return
			}
			if err := event.Client().Rest().PinMessage(event.ChannelID, event.MessageID); err != nil {
				slog.Error("error pinning poll message", slog.Any("error", err))
				return
			}
		}),
		bot.NewListenerFunc(func(event *events.GuildMessageUpdate) {
			if event.Message.Author.Bot || event.GuildID != p.DiscordConfig.GuildId {
				return
			}
			if event.Message.Poll != nil && event.Message.Poll.Results != nil && event.Message.Poll.Results.IsFinalized && event.Message.Pinned {
				if err := event.Client().Rest().UnpinMessage(event.ChannelID, event.MessageID); err != nil {
					slog.Error("error unpinning poll message", slog.Any("error", err))
					return
				}
			}
		}),
	)
}
