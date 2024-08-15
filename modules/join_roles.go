//go:build modules.all || modules.join_roles

package modules

import (
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
)

type JoinRolesConfig struct {
	UserRoleIds []snowflake.ID `env:"MODULES_JOIN_ROLES_USER_ROLES"`
	BotRoleIds  []snowflake.ID `env:"MODULES_JOIN_ROLES_BOT_ROLES"`
}

func init() {
	modules = append(modules, fx.Module("modules/join_roles",
		fx.Provide(
			env.ParseAs[JoinRolesConfig],
			ProvideJoinRoles,
		),
		fx.Invoke(NewJoinRoles),
	))

}

func ProvideJoinRoles() Results {
	return Results{
		Options: []bot.ConfigOpt{
			bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMembers)),
		},
	}
}

func NewJoinRoles(p ParamsWithConfig[JoinRolesConfig]) {
	p.Client.AddEventListeners(bot.NewListenerFunc(func(event *events.GuildMemberJoin) {
		var roles []snowflake.ID
		if event.Member.User.Bot {
			roles = p.Config.BotRoleIds
		} else {
			roles = p.Config.UserRoleIds
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
	}))
}
