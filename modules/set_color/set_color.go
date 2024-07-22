package set_color

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
	"github.com/v4violet/silly-club-bot/templates"
)

var setColorRegex = regexp.MustCompile("^#?([[:xdigit:]]{6})")

func Init() {
	modules.RegisterModule(modules.Module{
		Name: "set_color",
		Init: func() []bot.ConfigOpt {
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(onApplicationCommandInteractionCreate),
			}
		},
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
	})
}

func onApplicationCommandInteractionCreate(event *events.ApplicationCommandInteractionCreate) {
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
	color_raw := strings.ToLower(data.String("color"))
	var color int
	if color_raw == "random" {
		color = int(math.Round(rand.Float64() * 16777215))
	} else {
		color_string_matches := setColorRegex.FindStringSubmatch(color_raw)
		if len(color_string_matches) < 2 {
			event.CreateMessage(discord.MessageCreate{
				Content: templates.Exec("modules.set_color.errors.invalid_color", nil),
				Flags:   discord.MessageFlagEphemeral,
			})
			return
		}
		color_string := color_string_matches[1]
		maybe_color, err := strconv.ParseInt(color_string, 16, 32)
		if err != nil {
			slog.Error("error parsing color",
				slog.Any("error", err),
				slog.String("input", color_string),
			)
			event.CreateMessage(discord.MessageCreate{
				Content: templates.Exec("modules.set_color.errors.invalid_color", nil),
				Flags:   discord.MessageFlagEphemeral,
			})
			return
		}
		color = int(maybe_color)
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
			if role.ID == member_role_id && role.Name[:len("color/")] == "color/" {
				color_role = &role
				break out
			}
		}
	}
	color_str := strconv.FormatInt(int64(color), 16)
	role_name := fmt.Sprintf("color/#%s", color_str)
	permissions := discord.PermissionsNone
	if color_role != nil {
		if _, err := event.Client().Rest().UpdateRole(color_role.GuildID, color_role.ID, discord.RoleUpdate{
			Name:        &role_name,
			Color:       &color,
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
			Color:       color,
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
			Color: color,
		}).
		SetFlags(discord.MessageFlagEphemeral).
		Build(),
	)

	if config.Config.Modules.SetColor.LogChannelId == nil {
		return
	}

	now := time.Now()
	if _, err := event.Client().Rest().CreateMessage(*config.Config.Modules.SetColor.LogChannelId, discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Description: event.Member().Mention(),
				Color:       color,
				Timestamp:   &now,
			},
		},
	}); err != nil {
		slog.Error("error logging setcolor",
			slog.Any("error", err),
			slog.Any("channel_id", *config.Config.Modules.SetColor.LogChannelId),
		)
	}
}
