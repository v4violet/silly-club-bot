//go:build modules.all || modules.poll_pin

package modules

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"go.uber.org/fx"
)

func init() {
	modules = append(modules, fx.Module("modules/poll_pin",
		fx.Provide(ProvidePollPin),
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

func NewPollPin(p Params) {
	p.Client.AddEventListeners(
		bot.NewListenerFunc(func(event *events.GuildMessageCreate) {
			if event.Message.Author.Bot || event.GuildID != p.DiscordConfig.GuildId || event.Message.Poll == nil {
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
