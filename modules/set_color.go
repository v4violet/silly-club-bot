//go:build modules.all || modules.set_color

package modules

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

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

func colorToInt(color csscolorparser.Color) (int, error) {
	color_int, err := strconv.ParseInt(color.HexString()[1:7], 16, 32)
	return int(color_int), err
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
					color_raw := strings.TrimSpace(data.String("color"))
					var color csscolorparser.Color
					if strings.ToLower(color_raw) == "random" {
						color = csscolorparser.FromHsv(rand.Float64()*360, 0.6+(0.4*rand.Float64()), 1, 1)
					} else {
						parsed_color, err := csscolorparser.Parse(color_raw)
						if err != nil {
							event.CreateMessage(discord.MessageCreate{
								Content: err.Error(),
								Flags:   discord.MessageFlagEphemeral,
							})
							return
						}
						color = parsed_color
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
					color_str := color.HexString()
					color_int, err := colorToInt(color)
					if err != nil {
						event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
							SetContent(templates.Exec("modules.set_color.errors.color_convert", nil)).
							SetFlags(discord.MessageFlagEphemeral).
							Build(),
						)
						slog.Error("error converting color",
							slog.Any("error", err),
							slog.Any("guild_id", *event.GuildID()),
							slog.String("color", color.HexString()),
						)
						return
					}
					role_name := fmt.Sprintf("color/%s", color_str)
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
								"Color": color_str,
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
