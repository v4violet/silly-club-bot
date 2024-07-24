package join_roles

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
)

func Init() {
	modules.RegisterModule(modules.Module{
		Name: "join_roles",
		Init: func() []bot.ConfigOpt {
			if len(config.Config.Modules.JoinRoles.UserRoleIds) <= 0 && len(config.Config.Modules.JoinRoles.BotRoleIds) <= 0 {
				slog.Warn("`join_roles` module enabled but no user or bot role ids defined (skipping init)")
				return []bot.ConfigOpt{}
			}
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(onGuildMemberJoin),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMembers)),
			}
		},
	})
}

func onGuildMemberJoin(event *events.GuildMemberJoin) {
	var roles []snowflake.ID
	if event.Member.User.Bot {
		roles = config.Config.Modules.JoinRoles.BotRoleIds
	} else {
		roles = config.Config.Modules.JoinRoles.UserRoleIds
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
}
