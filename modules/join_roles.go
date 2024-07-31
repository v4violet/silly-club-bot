//go:build modules.all || modules.join_roles

package modules

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/sethvargo/go-envconfig"
)

var JoinRolesConfig struct {
	UserRoleIds []snowflake.ID `env:"MODULES_JOIN_ROLES_USER_ROLES"`
	BotRoleIds  []snowflake.ID `env:"MODULES_JOIN_ROLES_BOT_ROLES"`
}

func init() {
	Modules["join_roles"] = Module{
		Init: func() ([]bot.ConfigOpt, error) {
			snowflake.AllowUnquoted = true
			if err := envconfig.Process(context.Background(), &JoinRolesConfig); err != nil {
				return nil, err
			}

			if len(JoinRolesConfig.UserRoleIds) <= 0 && len(JoinRolesConfig.BotRoleIds) <= 0 {
				slog.Warn("`join_roles` module enabled but no user or bot role ids defined (skipping init)")
				return []bot.ConfigOpt{}, nil
			}
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(func(event *events.GuildMemberJoin) {
					var roles []snowflake.ID
					if event.Member.User.Bot {
						roles = JoinRolesConfig.BotRoleIds
					} else {
						roles = JoinRolesConfig.UserRoleIds
					}
					for _, roleId := range roles {
						if err := event.Client().Rest().AddMemberRole(event.GuildID, event.Member.User.ID, roleId); err != nil {
							slog.Error("error adding role",
								slog.Any("error", err),
								slog.Any("guild_id", event.GuildID),
								slog.Any("user_id", event.Member.User.ID),
								slog.Any("role_id", roleId),
							)
						}
					}
				}),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMembers)),
			}, nil
		},
	}
}
