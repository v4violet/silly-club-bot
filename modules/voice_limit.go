//go:build modules.all || modules.voice_limit

package modules

import (
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
)

type VoiceLimitConfig struct {
	Channel snowflake.ID `env:"MODULES_VOICE_LIMIT_CHANNEL,required,notEmpty"`
}

func init() {
	modules = append(modules, fx.Module("modules/voice_limig",
		fx.Provide(
			env.ParseAs[VoiceLimitConfig],
			ProvideVoiceLimit,
		),
		fx.Invoke(NewVoiceLimit),
	))
}

func ProvideVoiceLimit() Results {
	return Results{
		Options: []bot.ConfigOpt{
			bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuildMembers)),
		},
	}
}

func NewVoiceLimit(p ParamsWithConfig[VoiceLimitConfig]) {
	p.Client.AddEventListeners(
		bot.NewListenerFunc(func(event *events.GuildMemberJoin) { updateVoiceLimit(p) }),
		bot.NewListenerFunc(func(event *events.GuildMemberLeave) { updateVoiceLimit(p) }),
		bot.NewListenerFunc(func(event *events.GuildsReady) { updateVoiceLimit(p) }),
	)
}

func updateVoiceLimit(p ParamsWithConfig[VoiceLimitConfig]) {
	guild, found := p.Client.Caches().Guild(p.DiscordConfig.GuildId)
	if !found {
		return
	}
	if _, err := p.Client.Rest().UpdateChannel(p.Config.Channel, discord.GuildVoiceChannelUpdate{
		UserLimit: &guild.MemberCount,
	}); err != nil {
		slog.Error("error updating voice limit channel",
			slog.Any("error", err),
			slog.Any("channel_id", p.Config.Channel),
		)
	}
}
