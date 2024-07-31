//go:build modules.all || modules.user_log

package modules

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/sethvargo/go-envconfig"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/templates"
)

var UserLogConfig struct {
	Channel snowflake.ID `env:"MODULES_USER_LOG_CHANNEL,required"`
}

func init() {
	Modules["user_log"] = Module{
		Init: func() ([]bot.ConfigOpt, error) {
			snowflake.AllowUnquoted = true
			if err := envconfig.Process(context.Background(), &UserLogConfig); err != nil {
				return nil, err
			}

			return []bot.ConfigOpt{
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMembers)),
				bot.WithEventListenerFunc(func(event *events.GuildMemberJoin) {
					if config.Config.Discord.GuildId != event.GuildID {
						return
					}
					if _, err := event.Client().Rest().CreateMessage(UserLogConfig.Channel, discord.MessageCreate{
						Content: templates.Exec("modules.user_log.join", map[string]string{
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
				bot.WithEventListenerFunc(func(event *events.GuildMemberLeave) {
					if config.Config.Discord.GuildId != event.GuildID {
						return
					}
					if _, err := event.Client().Rest().CreateMessage(UserLogConfig.Channel, discord.MessageCreate{
						Content: templates.Exec("modules.user_log.leave", map[string]string{
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
			}, nil
		},
	}
}
