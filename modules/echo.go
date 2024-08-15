//go:build modules.all || modules.echo

package modules

import (
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"go.uber.org/fx"
)

func init() {
	modules = append(modules, fx.Module("modules/echo",
		fx.Provide(ProvideEcho),
		fx.Invoke(NewEcho),
	))
}

func ProvideEcho() ResultsWithApplicationCommands {
	return ResultsWithApplicationCommands{
		ApplicationCommand: []discord.ApplicationCommandCreate{
			discord.SlashCommandCreate{
				Name:        "echo",
				Description: "echo text",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "text",
						Description: "the text to echo",
						Required:    true,
					},
				},
			},
		},
		Options: []bot.ConfigOpt{
			bot.WithGatewayConfigOpts(
				gateway.WithIntents(gateway.IntentGuildMessages, gateway.IntentMessageContent),
			),
		},
	}
}

func NewEcho(p Params) {
	p.Client.AddEventListeners(
		bot.NewListenerFunc(func(event *events.GuildMessageCreate) {
			if event.Message.Author.Bot || event.GuildID != p.DiscordConfig.GuildId {
				return
			}
			content := strings.TrimSpace(event.Message.Content)
			if !strings.HasPrefix(strings.ToLower(content), "!echo") {
				return
			}
			event.Client().Rest().CreateMessage(event.ChannelID, discord.MessageCreate{
				Content: content[5:],
				MessageReference: &discord.MessageReference{
					GuildID:   &event.GuildID,
					ChannelID: &event.ChannelID,
					MessageID: &event.MessageID,
				},
				AllowedMentions: &discord.AllowedMentions{
					Parse:       []discord.AllowedMentionType{},
					RepliedUser: true,
				},
			})
		}),
		bot.NewListenerFunc(func(event *events.ApplicationCommandInteractionCreate) {
			if event.GuildID() == nil || event.Data.CommandName() != "echo" {
				return
			}
			event.CreateMessage(discord.MessageCreate{
				Content: event.SlashCommandInteractionData().String("text"),
				AllowedMentions: &discord.AllowedMentions{
					Parse:       []discord.AllowedMentionType{},
					RepliedUser: true,
				},
			})
		}),
	)
}
