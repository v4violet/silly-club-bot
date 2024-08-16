//go:build modules.all || modules.set_color

package modules

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/mazznoer/csscolorparser"
	"github.com/v4violet/silly-club-bot/templateutils"

	"github.com/caarlos0/env/v11"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
)

var validColorNames = []string{
	"default",
	"random",

	"aliceblue",
	"antiquewhite",
	"aqua",
	"aquamarine",
	"azure",
	"beige",
	"bisque",
	"black",
	"blanchedalmond",
	"blue",
	"blueviolet",
	"brown",
	"burlywood",
	"cadetblue",
	"chartreuse",
	"chocolate",
	"coral",
	"cornflowerblue",
	"cornsilk",
	"crimson",
	"cyan",
	"darkblue",
	"darkcyan",
	"darkgoldenrod",
	"darkgray",
	"darkgreen",
	"darkgrey",
	"darkkhaki",
	"darkmagenta",
	"darkolivegreen",
	"darkorange",
	"darkorchid",
	"darkred",
	"darksalmon",
	"darkseagreen",
	"darkslateblue",
	"darkslategray",
	"darkslategrey",
	"darkturquoise",
	"darkviolet",
	"deeppink",
	"deepskyblue",
	"dimgray",
	"dimgrey",
	"dodgerblue",
	"firebrick",
	"floralwhite",
	"forestgreen",
	"fuchsia",
	"gainsboro",
	"ghostwhite",
	"gold",
	"goldenrod",
	"gray",
	"green",
	"greenyellow",
	"grey",
	"honeydew",
	"hotpink",
	"indianred",
	"indigo",
	"ivory",
	"khaki",
	"lavender",
	"lavenderblush",
	"lawngreen",
	"lemonchiffon",
	"lightblue",
	"lightcoral",
	"lightcyan",
	"lightgoldenrodyellowlightgray",
	"lightgreen",
	"lightgrey",
	"lightpink",
	"lightsalmon",
	"lightseagreen",
	"lightskyblue",
	"lightslategray",
	"lightslategrey",
	"lightsteelblue",
	"lightyellow",
	"lime",
	"limegreen",
	"linen",
	"magenta",
	"maroon",
	"mediumaquamarine",
	"mediumblue",
	"mediumorchid",
	"mediumpurple",
	"mediumseagreen",
	"mediumslateblue",
	"mediumspringgreen",
	"mediumturquoise",
	"mediumvioletred",
	"midnightblue",
	"mintcream",
	"mistyrose",
	"moccasin",
	"navajowhite",
	"navy",
	"oldlace",
	"olive",
	"olivedrab",
	"orange",
	"orangered",
	"orchid",
	"palegoldenrod",
	"palegreen",
	"paleturquoise",
	"palevioletred",
	"papayawhip",
	"peachpuff",
	"peru",
	"pink",
	"plum",
	"powderblue",
	"purple",
	"rebeccapurple",
	"red",
	"rosybrown",
	"royalblue",
	"saddlebrown",
	"salmon",
	"sandybrown",
	"seagreen",
	"seashell",
	"sienna",
	"silver",
	"skyblue",
	"slateblue",
	"slategray",
	"slategrey",
	"snow",
	"springgreen",
	"steelblue",
	"tan",
	"teal",
	"thistle",
	"tomato",
	"turquoise",
	"violet",
	"wheat",
	"white",
	"whitesmoke",
	"yellow",
	"yellowgreen",
}

type SetColorConfig struct {
	LogChannel snowflake.ID `env:"MODULES_SET_COLOR_LOG_CHANNEL,required,notEmpty"`
}

func init() {
	modules = append(modules, fx.Module("modules/set_color",
		fx.Provide(env.ParseAs[SetColorConfig], ProvideSetColor),
		fx.Invoke(NewSetColor),
	))
}

func ProvideSetColor() ApplicationCommandsResults {
	return ApplicationCommandsResults{
		ApplicationCommand: []discord.ApplicationCommandCreate{
			discord.SlashCommandCreate{
				Name:        "setcolor",
				Description: "set your custom role color",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:         "color",
						Description:  "css color",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
		},
	}
}

func colorToInt(color csscolorparser.Color, allow_black bool) (int, error) {
	color_int, err := strconv.ParseInt(color.HexString()[1:7], 16, 32)
	if color_int == 0 && !allow_black {
		color_int = 0x000100
	}
	return int(color_int), err
}

func randomColor() csscolorparser.Color {
	return csscolorparser.FromHsv(rand.Float64()*360, 0.6+(0.4*rand.Float64()), 1, 1)
}

func NewSetColor(p ParamsWithConfigAndTemplate[SetColorConfig]) {
	p.Client.AddEventListeners(
		bot.NewListenerFunc(func(event *events.AutocompleteInteractionCreate) {
			if event.GuildID() == nil || event.Data.CommandName != "setcolor" || event.Data.Focused().Name != "color" {
				return
			}
			if *event.GuildID() != p.DiscordConfig.GuildId {
				return
			}

			choices := []discord.AutocompleteChoice{}

			color := strings.ToLower(strings.TrimSpace(event.Data.String("color")))

			if len(color) > 1 {
				matches := fuzzy.RankFind(color, validColorNames)
				sort.Sort(matches)
				if len(matches) > 0 {
					if matches[0].Target != color {
						choices = append(choices, discord.AutocompleteChoiceString{
							Name:  color,
							Value: color,
						})
					}
					for _, match := range matches {
						choices = append(choices, discord.AutocompleteChoiceString{
							Name:  match.Target,
							Value: match.Target,
						})
					}
				}
			}

			if len(choices) <= 0 {
				choices = append(choices, discord.AutocompleteChoiceString{
					Name:  color,
					Value: color,
				}, discord.AutocompleteChoiceString{
					Name:  "random",
					Value: "random",
				}, discord.AutocompleteChoiceString{
					Name:  "default",
					Value: "default",
				})
			}

			event.AutocompleteResult(choices)
		}),
		bot.NewListenerFunc(func(event *events.ApplicationCommandInteractionCreate) {
			if event.GuildID() == nil || event.Data.CommandName() != "setcolor" {
				return
			}
			if *event.GuildID() != p.DiscordConfig.GuildId {
				return
			}

			data := event.SlashCommandInteractionData()
			color_raw := strings.ToLower(strings.TrimSpace(data.String("color")))
			allow_black := false
			var color csscolorparser.Color
			if color_raw == "random" {
				color = randomColor()
			} else if color_raw == "default" {
				allow_black = true
				color = csscolorparser.Color{R: 0, G: 0, B: 0, A: 0}
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
					SetContent(templateutils.MustExecuteTemplateToString(p.Template, "modules.set_color.errors.guild_roles", nil)).
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
			if allow_black {
				color_str = "default"
			}
			color_int, err := colorToInt(color, allow_black)
			if err != nil {
				event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
					SetContent("").
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
						SetContent(templateutils.MustExecuteTemplateToString(p.Template, "modules.set_color.errors.role_update", nil)).
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
						SetContent(templateutils.MustExecuteTemplateToString(p.Template, "modules.set_color.errors.role_create", nil)).
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
						SetContent(templateutils.MustExecuteTemplateToString(p.Template, "modules.set_color.errors.role_add_member", nil)).
						SetFlags(discord.MessageFlagEphemeral).
						Build(),
					)
					return
				}
			}

			event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdateBuilder().
				SetEmbeds(discord.Embed{
					Title: templateutils.MustExecuteTemplateToString(p.Template, "modules.set_color.success", map[string]string{
						"Color": color_str,
					}),
					Color: color_int,
				}).
				SetFlags(discord.MessageFlagEphemeral).
				Build(),
			)

			now := time.Now()
			event.Client().Rest().CreateMessage(p.Config.LogChannel, discord.MessageCreate{
				Embeds: []discord.Embed{
					{
						Description: event.Member().Mention(),
						Color:       color_int,
						Timestamp:   &now,
					},
				},
			})
		}),
	)
}
