//go:build modules.all || modules.random_react

package modules

import (
	"context"
	"log/slog"
	"math/rand/v2"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/sethvargo/go-envconfig"
	"github.com/v4violet/silly-club-bot/config"
)

var RandomReactConfig struct {
	Percentage float64 `env:"MODULES_RANDOM_REACT_PERCENTAGE,default=0.01"`
	Emoji      string  `env:"MODULES_RANDOM_REACT_EMOJI,required"`

	LogChannel snowflake.ID `env:"MODULES_RANDOM_REACT_LOG_CHANNEL,required"`

	WhitelistedChannelsRaw []snowflake.ID `env:"MODULES_RANDOM_REACT_WHITELISTED_CHANNELS"`
	WhitelistedChannelIds  map[snowflake.ID]bool
}

func init() {
	Modules["random_react"] = Module{
		Init: func() ([]bot.ConfigOpt, error) {
			snowflake.AllowUnquoted = true
			if err := envconfig.Process(context.Background(), &RandomReactConfig); err != nil {
				return nil, err
			}

			RandomReactConfig.WhitelistedChannelIds = make(map[snowflake.ID]bool)

			for _, channel := range RandomReactConfig.WhitelistedChannelsRaw {
				RandomReactConfig.WhitelistedChannelIds[channel] = true
			}

			return []bot.ConfigOpt{
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessages)),
				bot.WithEventListenerFunc(func(event *events.GuildMessageCreate) {
					if event.Message.Author.Bot || event.GuildID != config.Config.Discord.GuildId {
						return
					}

					if _, channelWhitelisted := RandomReactConfig.WhitelistedChannelIds[event.ChannelID]; !channelWhitelisted {
						return
					}

					if rand.Float64() > (RandomReactConfig.Percentage / 100) {
						return
					}

					if err := event.Client().Rest().AddReaction(event.ChannelID, event.MessageID, RandomReactConfig.Emoji); err != nil {
						slog.Error("error adding reaction",
							slog.Any("error", err),
							slog.Any("channel_id", event.ChannelID),
							slog.Any("message_id", event.MessageID),
						)
						return
					}
					if _, err := event.Client().Rest().CreateMessage(RandomReactConfig.LogChannel, discord.MessageCreate{
						Content: event.Message.JumpURL(),
					}); err != nil {
						slog.Error("error logging random reaction",
							slog.Any("error", err),
							slog.Any("channel_id", event.ChannelID),
							slog.Any("message_id", event.MessageID),
						)
					}
				}),
			}, nil
		},
	}
}
