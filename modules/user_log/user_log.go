package user_log

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
	"github.com/v4violet/silly-club-bot/templates"
)

func Init() {
	modules.RegisterModule(modules.Module{
		Name: "user_log",
		Init: func() []bot.ConfigOpt {
			if config.Config.Modules.UserLog.ChannelId == nil {
				slog.Warn("`user_log` module enabled but missing channel id (skipping init)")
				return []bot.ConfigOpt{}
			}
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(onGuildMemberJoin),
				bot.WithEventListenerFunc(onGuildMemberLeave),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMembers)),
			}
		},
	})
}

func onGuildMemberJoin(event *events.GuildMemberJoin) {
	if config.Config.Discord.GuildId != event.GuildID {
		return
	}
	if config.Config.Modules.UserLog.ChannelId == nil {
		return
	}
	if _, err := event.Client().Rest().CreateMessage(*config.Config.Modules.UserLog.ChannelId, discord.MessageCreate{
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
}

func onGuildMemberLeave(event *events.GuildMemberLeave) {
	if config.Config.Discord.GuildId != event.GuildID {
		return
	}
	if config.Config.Modules.UserLog.ChannelId == nil {
		return
	}
	if _, err := event.Client().Rest().CreateMessage(*config.Config.Modules.UserLog.ChannelId, discord.MessageCreate{
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
}
