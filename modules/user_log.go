//go:build modules.all || modules.user_log

package modules

import (
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/v4violet/silly-club-bot/templateutils"
	"go.uber.org/fx"
)

type UserLogConfig struct {
	LogChannel snowflake.ID `env:"MODULES_USER_LOG_CHANNEL,required,notEmpty"`
}

func init() {
	modules = append(modules, fx.Module("modules/user_log",
		fx.Provide(
			env.ParseAs[UserLogConfig],
			ProvideUserLog,
		),
		fx.Invoke(NewUserLog),
	))
}

func ProvideUserLog() Results {
	return Results{
		Options: []bot.ConfigOpt{
			bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMembers)),
		},
	}
}

func NewUserLog(p ParamsWithConfigAndTemplate[UserLogConfig]) {
	p.Client.AddEventListeners(
		bot.NewListenerFunc(func(event *events.GuildMemberJoin) {
			if p.DiscordConfig.GuildId != event.GuildID {
				return
			}
			if _, err := event.Client().Rest().CreateMessage(p.Config.LogChannel, discord.MessageCreate{
				Content: templateutils.MustExecuteTemplateToString(p.Template, "modules.user_log.join", map[string]string{
					"User": event.Member.Mention(),
				}),
				AllowedMentions: &discord.AllowedMentions{
					Parse: []discord.AllowedMentionType{},
				},
			}); err != nil {
				slog.Error("error logging user join", slog.Any("error", err),
					slog.Any("guild_id", event.GuildID),
					slog.Any("user_id", event.Member.User.ID),
				)
			}
		}),
		bot.NewListenerFunc(func(event *events.GuildMemberLeave) {
			if p.DiscordConfig.GuildId != event.GuildID {
				return
			}
			if _, err := event.Client().Rest().CreateMessage(p.Config.LogChannel, discord.MessageCreate{
				Content: templateutils.MustExecuteTemplateToString(p.Template, "modules.user_log.leave", map[string]string{
					"User": event.Member.Mention(),
				}),
				AllowedMentions: &discord.AllowedMentions{
					Parse: []discord.AllowedMentionType{},
				},
			}); err != nil {
				slog.Error("error logging user join", slog.Any("error", err),
					slog.Any("guild_id", event.GuildID),
					slog.Any("user_id", event.Member.User.ID),
				)
			}
		}),
	)
}
