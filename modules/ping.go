//go:build modules.all || modules.ping

package modules

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
)

func init() {
	Modules["ping"] = Module{
		Init: func() ([]bot.ConfigOpt, error) {
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(func(event *events.GuildMessageCreate) {
					if event.Message.Author.Bot || event.GuildID != config.Config.Discord.GuildId {
						return
					}
					if !strings.HasPrefix(strings.ToLower(event.Message.Content), "!ping") {
						return
					}

					inline := true

					gateway_latency := fmt.Sprintf("`%s`", event.Client().Gateway().Latency().Round(10*time.Microsecond).String())
					rest_latency_start := time.Now()
					message, err := event.Client().Rest().CreateMessage(event.ChannelID, discord.MessageCreate{
						Embeds: []discord.Embed{
							discord.NewEmbedBuilder().
								SetTitle("Pong!").
								SetFields(
									discord.EmbedField{
										Name:   "📡 Gateway",
										Value:  gateway_latency,
										Inline: &inline,
									},
									discord.EmbedField{
										Name:   "💬 API",
										Value:  "Loading...",
										Inline: &inline,
									},
								).
								SetColor(0xffff00).
								Build(),
						},
						MessageReference: &discord.MessageReference{
							MessageID:       &event.MessageID,
							ChannelID:       &event.ChannelID,
							GuildID:         &event.GuildID,
							FailIfNotExists: false,
						},
					})
					rest_latency := fmt.Sprintf("`%s`", time.Since(rest_latency_start).Round(10*time.Microsecond).String())
					if err != nil {
						slog.Error("error sending pong",
							slog.Any("error", err),
							slog.Any("channel_id", event.ChannelID),
							slog.Any("message_id", event.MessageID),
						)
						return
					}

					if _, err := event.Client().Rest().UpdateMessage(event.ChannelID, message.ID, discord.MessageUpdate{
						Embeds: &[]discord.Embed{
							discord.NewEmbedBuilder().
								SetTitle("Pong!").
								SetFields(
									discord.EmbedField{
										Name:   "📡 Gateway",
										Value:  gateway_latency,
										Inline: &inline,
									},
									discord.EmbedField{
										Name:   "💬 API",
										Value:  rest_latency,
										Inline: &inline,
									},
								).
								SetColor(0x00ff00).
								Build(),
						},
					}); err != nil {
						slog.Error("error updating pong",
							slog.Any("error", err),
							slog.Any("channel_id", message.ChannelID),
							slog.Any("message_id", message.ID),
						)
					}
				}),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMessages, gateway.IntentMessageContent)),
			}, nil
		},
	}
}
