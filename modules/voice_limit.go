//go:build modules.all || modules.voice_limit

package modules

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/sethvargo/go-envconfig"
	"github.com/v4violet/silly-club-bot/config"
)

var VoiceLimitConfig struct {
	Channel snowflake.ID `env:"MODULES_VOICE_LIMIT_CHANNEL,required"`
}

func init() {
	Modules["voice_limit"] = Module{
		Init: func() ([]bot.ConfigOpt, error) {
			snowflake.AllowUnquoted = true
			if err := envconfig.Process(context.Background(), &VoiceLimitConfig); err != nil {
				return nil, err
			}

			return []bot.ConfigOpt{
				bot.WithEventListenerFunc(func(event *events.GuildMemberJoin) { updateVoiceLimit(event.Client()) }),
				bot.WithEventListenerFunc(func(event *events.GuildMemberLeave) { updateVoiceLimit(event.Client()) }),
				bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMembers)),
				bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds)),
			}, nil
		},
	}
}

func updateVoiceLimit(client bot.Client) {
	guild, found := client.Caches().Guild(config.Config.Discord.GuildId)
	if !found {
		return
	}
	if _, err := client.Rest().UpdateChannel(VoiceLimitConfig.Channel, discord.GuildVoiceChannelUpdate{
		UserLimit: &guild.MemberCount,
	}); err != nil {
		slog.Error("error updating voice limit channel",
			slog.Any("error", err),
			slog.Any("channel_id", VoiceLimitConfig.Channel),
		)
	}
}
