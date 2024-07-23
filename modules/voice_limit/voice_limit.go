package voice_limit

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
)

func Init() {
	modules.RegisterModule(modules.Module{
		Name: "voice_limit",
		Init: func() []bot.ConfigOpt {
			if config.Config.Modules.UserLog.ChannelId == nil {
				slog.Warn("`voice_limit` module enabled but missing channel id (skipping init)")
				return []bot.ConfigOpt{}
			}
			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(func(event *events.GuildMemberJoin) { updateVoiceLimit(event.Client()) }),
				bot.WithEventListenerFunc(func(event *events.GuildMemberLeave) { updateVoiceLimit(event.Client()) }),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMembers)),
				bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds)),
			}
		},
	})
}

func updateVoiceLimit(client bot.Client) {
	if config.Config.Modules.VoiceLimit.ChannelId == nil {
		return
	}
	guild, found := client.Caches().Guild(config.Config.Discord.GuildId)
	if !found {
		return
	}
	if _, err := client.Rest().UpdateChannel(*config.Config.Modules.VoiceLimit.ChannelId, discord.GuildVoiceChannelUpdate{
		UserLimit: &guild.MemberCount,
	}); err != nil {
		slog.Error("error updating voice limit channel",
			slog.Any("error", err),
			slog.Any("channel_id", *config.Config.Modules.VoiceLimit.ChannelId),
		)
	}
}
