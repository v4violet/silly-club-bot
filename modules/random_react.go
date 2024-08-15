//go:build modules.all || modules.random_react

package modules

import (
	"log/slog"
	"math/rand/v2"

	"github.com/caarlos0/env/v11"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
)

type RandomReactConfig struct {
	Percentage float64 `env:"MODULES_RANDOM_REACT_PERCENTAGE" envDefault:"0.01"`
	Emoji      string  `env:"MODULES_RANDOM_REACT_EMOJI,required,notEmpty"`

	LogChannel snowflake.ID `env:"MODULES_RANDOM_REACT_LOG_CHANNEL,required,notEmpty"`

	WhitelistedChannelsRaw []snowflake.ID `env:"MODULES_RANDOM_REACT_WHITELISTED_CHANNELS"`
	WhitelistedChannelIds  map[snowflake.ID]bool
}

func init() {
	modules = append(modules, fx.Module("modules/random_react",
		fx.Provide(
			env.ParseAs[RandomReactConfig],
			ProvideRandomReact,
		),
		fx.Invoke(NewRandomReact),
	))
}

func ProvideRandomReact() Results {
	return Results{
		Options: []bot.ConfigOpt{
			bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessages)),
		},
	}
}

func NewRandomReact(p ParamsWithConfig[RandomReactConfig]) {
	p.Config.WhitelistedChannelIds = make(map[snowflake.ID]bool)

	for _, channel := range p.Config.WhitelistedChannelsRaw {
		p.Config.WhitelistedChannelIds[channel] = true
	}
	p.Client.AddEventListeners(bot.NewListenerFunc(func(event *events.GuildMessageCreate) {
		if event.Message.Author.Bot || event.GuildID != p.DiscordConfig.GuildId {
			return
		}

		if _, channelWhitelisted := p.Config.WhitelistedChannelIds[event.ChannelID]; !channelWhitelisted {
			return
		}

		if rand.Float64() > (p.Config.Percentage / 100) {
			return
		}

		if err := event.Client().Rest().AddReaction(event.ChannelID, event.MessageID, p.Config.Emoji); err != nil {
			slog.Error("error adding reaction",
				slog.Any("error", err),
				slog.Any("channel_id", event.ChannelID),
				slog.Any("message_id", event.MessageID),
			)
			return
		}
		if _, err := event.Client().Rest().CreateMessage(p.Config.LogChannel, discord.MessageCreate{
			Content: event.Message.JumpURL(),
		}); err != nil {
			slog.Error("error logging random reaction",
				slog.Any("error", err),
				slog.Any("channel_id", event.ChannelID),
				slog.Any("message_id", event.MessageID),
			)
		}
	}))
}
