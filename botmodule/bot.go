package botmodule

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"go.uber.org/fx"
)

type DiscordConfig struct {
	Token   string       `env:"DISCORD_TOKEN,required,notEmpty"`
	GuildId snowflake.ID `env:"DISCORD_GUILD_ID,required,notEmpty"`
}

type DryRun struct {
	DryRun bool
}

var DefaultOptions = []bot.ConfigOpt{
	bot.WithGatewayConfigOpts(
		gateway.WithAutoReconnect(true),
		gateway.WithCompress(true),
		gateway.WithIntents(gateway.IntentGuilds),
	),
	bot.WithCacheConfigOpts(
		cache.WithCaches(cache.FlagGuilds),
	),
	bot.WithEventManagerConfigOpts(
		bot.WithAsyncEventsEnabled(),
	),
}

var Module = fx.Module("bot",
	fx.Supply(fx.Annotate(DefaultOptions, fx.ResultTags(`group:"bot,flatten"`))),
	fx.Provide(
		env.ParseAs[DiscordConfig],
		NewBot,
	),
)

type Params struct {
	fx.In

	Options             []bot.ConfigOpt                    `group:"bot"`
	ApplicationCommands []discord.ApplicationCommandCreate `group:"bot"`
	Config              DiscordConfig
	DryRun              DryRun
	LC                  fx.Lifecycle
}

var commit = func() string {
	buildinfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, kv := range buildinfo.Settings {
			if kv.Key == "vcs.revision" {
				return kv.Value
			}
		}
	}
	return ""
}()

func setStatus(client bot.Client) {
	activity := client.Gateway().Latency().Round(time.Microsecond * 10).String()
	if len(commit) != 0 {
		activity = fmt.Sprintf("%s | %s", commit[:7], activity)
	}
	if err := client.SetPresence(context.Background(), gateway.WithCustomActivity(activity)); err != nil {
		slog.Warn("error setting presence", slog.Any("error", err))
	}
}

func NewBot(p Params) (bot.Client, error) {
	p.Options = append(p.Options,
		bot.WithEventListenerFunc(func(event *events.Ready) {
			slog.Info("ready",
				slog.String("user_id", event.User.ID.String()),
				slog.String("user_tag", event.User.Tag()),
			)
			setStatus(event.Client())
		}),
		bot.WithEventListenerFunc(func(event *events.GuildsReady) {
			slog.Info("guilds ready", slog.Int("count", event.Client().Caches().GuildsLen()))
		}),
		bot.WithEventListenerFunc(func(event *events.Resumed) {
			slog.Info("resumed", slog.Int("sequence", event.SequenceNumber()))
			setStatus(event.Client())
		}),
		bot.WithEventListenerFunc(func(event *events.HeartbeatAck) {
			//latency calculation happens after event listener calls for some reason?
			go func() {
				time.Sleep(time.Millisecond * 10)
				setStatus(event.Client())
			}()
		}),
	)
	client, err := disgo.New(p.Config.Token, p.Options...)
	if err != nil {
		return nil, err
	}

	if !p.DryRun.DryRun {
		slog.Info("starting",
			slog.Any("intents", client.Gateway().Intents()),
			slog.Any("caches", client.Caches().CacheFlags()),
			slog.Any("guild_id", p.Config.GuildId),
		)

		if _, err := client.Rest().SetGuildCommands(client.ApplicationID(), p.Config.GuildId, p.ApplicationCommands); err != nil {
			return nil, err
		}

		slog.Info("set guild commands", slog.Int("command_count", len(p.ApplicationCommands)))

		p.LC.Append(fx.StartStopHook(client.OpenGateway, func(ctx context.Context) {
			slog.Info("shutting down")
			client.SetPresence(ctx, gateway.WithOnlineStatus(discord.OnlineStatusInvisible))
			client.Close(ctx)
			slog.Info("goodbye")
		}))
	}

	return client, nil
}
