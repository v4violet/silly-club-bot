//go:build modules.all || modules.set_color

package modules

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/crazy3lf/colorconv"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/mazznoer/csscolorparser"
	"github.com/sethvargo/go-envconfig"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/templates"
)

var SetColorConfig struct {
	LogChannel snowflake.ID `env:"MODULES_SET_COLOR_LOG_CHANNEL,required"`
}

func rgbToInt(r, g, b uint8) int {
	return (0xFFFF * int(r)) + (0xFF * int(g)) + int(b)
}

func init() {
	Modules["set_color"] = Module{
		ApplicationCommands: &[]discord.ApplicationCommandCreate{
			discord.SlashCommandCreate{
				Name:        "setcolor",
				Description: "set your custom role color",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "color",
						Description: "hex color",
						Required:    true,
					},
				},
			},
		},
		Init: func() ([]bot.ConfigOpt, error) {
			snowflake.AllowUnquoted = true
			if err := envconfig.Process(context.Background(), &SetColorConfig); err != nil {
				return nil, err
			}

			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(func(event *events.ApplicationCommandInteractionCreate) {
					if event.GuildID() == nil || event.Data.CommandName() != "setcolor" {
						return
					}
					if *event.GuildID() != config.Config.Discord.GuildId {
						slog.Warn("this isnt our guild, ignoring",
							slog.Any("expected", event.GuildID()),
							slog.Any("found", config.Config.Discord.GuildId),
						)
						return
					}

					data := event.SlashCommandInteractionData()
					color_raw := data.String("color")
					var color_int int
					if strings.ToLower(color_raw) == "random" {
						// generate random hue + saturation between 60% and 100%
						r, g, b, err := colorconv.HSVToRGB(rand.Float64()*360, 0.6+(0.4*rand.Float64()), 1)
						if err != nil {
							event.CreateMessage(discord.MessageCreate{
								Content: templates.Exec("modules.set_color.errors.random_color", nil),
								Flags:   discord.MessageFlagEphemeral,
							})
							slog.Error("error converting random color", slog.Any("error", err))
							return
						}
						color_int = rgbToInt(r, g, b)
					} else {
						color, err := csscolorparser.Parse(color_raw)
						if err != nil {
							event.CreateMessage(discord.MessageCreate{
								Content: templates.Exec("modules.set_color.errors.invalid_color", nil),
								Flags:   discord.MessageFlagEphemeral,
							})
							return
						}
						r, g, b, _ := color.RGBA255()
						color_int = rgbToInt(r, g, b)
					}

					event.DeferCreateMessage(true)

					guild_roles, err := event.Client().Rest().GetRoles(*event.GuildID())
					if err != nil {
						event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
							SetContent(templates.Exec("modules.set_color.errors.guild_roles", nil)).
							SetFlags(discord.MessageFlagEphemeral).
							Build(),
						)
						slog.Error("error getting guild roles",
							slog.Any("error", err),
							slog.Any("guild_id", *event.GuildID()),
						)
						return
					}
					var color_role *discord.Role
				out:
					for _, role := range guild_roles {
						for _, member_role_id := range event.Member().RoleIDs {
							if role.ID == member_role_id && strings.HasPrefix(role.Name, "color/") {
								color_role = &role
								break out
							}
						}
					}
					color_str := fmt.Sprintf("%06x", color_int)
					role_name := fmt.Sprintf("color/#%s", color_str)
					permissions := discord.PermissionsNone
					if color_role != nil {
						if _, err := event.Client().Rest().UpdateRole(color_role.GuildID, color_role.ID, discord.RoleUpdate{
							Name:        &role_name,
							Color:       &color_int,
							Permissions: &permissions,
						}); err != nil {
							event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
								SetContent(templates.Exec("modules.set_color.errors.role_update", nil)).
								SetFlags(discord.MessageFlagEphemeral).
								Build(),
							)
							slog.Error("error updating color role",
								slog.Any("error", err),
								slog.Any("guild_id", *event.GuildID()),
								slog.Any("role_id", color_role.ID),
							)
							return
						}
					} else {
						role, err := event.Client().Rest().CreateRole(*event.GuildID(), discord.RoleCreate{
							Name:        role_name,
							Color:       color_int,
							Permissions: &permissions,
						})
						if err != nil {
							slog.Error("error creating color role",
								slog.Any("error", err),
								slog.Any("guild_id", *event.GuildID()),
							)
							event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
								SetContent(templates.Exec("modules.set_color.errors.role_create", nil)).
								SetFlags(discord.MessageFlagEphemeral).
								Build(),
							)
							return
						}
						if err := event.Client().Rest().AddMemberRole(*event.GuildID(), event.User().ID, role.ID); err != nil {
							slog.Error("error adding color role to member",
								slog.Any("error", err),
								slog.Any("guild_id", *event.GuildID()),
								slog.Any("role_id", role.ID),
								slog.Any("user_id", event.User().ID),
							)
							event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
								SetContent(templates.Exec("modules.set_color.errors.role_add_member", nil)).
								SetFlags(discord.MessageFlagEphemeral).
								Build(),
							)
							return
						}
					}

					event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
						SetEmbeds(discord.Embed{
							Title: templates.Exec("modules.set_color.success", map[string]string{
								"Color": "#" + color_str,
							}),
							Color: color_int,
						}).
						SetFlags(discord.MessageFlagEphemeral).
						Build(),
					)

					now := time.Now()
					if _, err := event.Client().Rest().CreateMessage(SetColorConfig.LogChannel, discord.MessageCreate{
						Embeds: []discord.Embed{
							{
								Description: event.Member().Mention(),
								Color:       color_int,
								Timestamp:   &now,
							},
						},
					}); err != nil {
						slog.Error("error logging setcolor",
							slog.Any("error", err),
							slog.Any("channel_id", SetColorConfig.LogChannel),
						)
					}
				}),
			}, nil
		},
	}
}
