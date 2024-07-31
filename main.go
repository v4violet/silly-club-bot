package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/build"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
	"github.com/v4violet/silly-club-bot/templates"
	_ "go.uber.org/automaxprocs"
)

func createStatus(latency *time.Duration) gateway.PresenceOpt {
	if latency != nil {
		l := latency.Round(time.Microsecond * 10)
		latency = &l
	}
	return gateway.WithCustomActivity(templates.Exec("generic.status", map[string]any{
		"Version": build.Version,
		"Latency": latency,
	}))
}

func main() {
	bot_config := []bot.ConfigOpt{
		bot.WithGatewayConfigOpts(
			gateway.WithCompress(true),
			gateway.WithAutoReconnect(true),
			gateway.WithLogger(slog.Default()),
			gateway.WithIntents(gateway.IntentGuilds),
			gateway.WithPresenceOpts(createStatus(nil)),
		),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds)),
		bot.WithEventListenerFunc(func(event *events.Ready) {
			slog.Info("ready",
				slog.String("user_id", event.User.ID.String()),
				slog.String("user_tag", event.User.Tag()),
			)
			latency := event.Client().Gateway().Latency()
			if err := event.Client().SetPresence(context.Background(), createStatus(&latency)); err != nil {
				slog.Warn("error setting presence", slog.Any("error", err))
			}
		}),
		bot.WithEventListenerFunc(func(event *events.GuildsReady) {
			slog.Info("guilds ready", slog.Int("count", event.Client().Caches().GuildsLen()))
		}),
		bot.WithEventListenerFunc(func(event *events.Resumed) {
			slog.Info("resumed", slog.Int("sequence", event.SequenceNumber()))
			latency := event.Client().Gateway().Latency()
			if err := event.Client().SetPresence(context.Background(), createStatus(&latency)); err != nil {
				slog.Warn("error setting presence", slog.Any("error", err))
			}
		}),
		bot.WithEventListenerFunc(func(event *events.HeartbeatAck) {
			//latency calculation happens after event listener calls for some reason?
			go func() {
				time.Sleep(time.Millisecond * 10)
				latency := event.Client().Gateway().Latency()
				if err := event.Client().SetPresence(context.Background(), createStatus(&latency)); err != nil {
					slog.Warn("error setting presence", slog.Any("error", err))
				}
			}()
		}),
	}

	application_commands := []discord.ApplicationCommandCreate{}

	enabled_modules := []string{}

	for k, v := range modules.Modules {
		enabled_modules = append(enabled_modules, k)
		module_config, err := v.Init()
		if err != nil {
			slog.Error("error initializing module", slog.String("module", k), slog.Any("error", err))
			os.Exit(1)
		}
		bot_config = append(bot_config, module_config...)
		if v.ApplicationCommands != nil {
			application_commands = append(application_commands, *v.ApplicationCommands...)
		}
	}

	client, err := disgo.New(config.Config.Discord.Token, bot_config...)

	if err != nil {
		slog.Error("error constructing bot client", slog.Any("error", err))
		os.Exit(1)
		return
	}

	slog.Info("starting",
		slog.String("version", build.Version),
		slog.Any("modules", strings.Join(enabled_modules, ",")),
		slog.Any("intents", client.Gateway().Intents()),
		slog.Any("caches", client.Caches().CacheFlags()),
		slog.Any("guild_id", config.Config.Discord.GuildId),
	)

	if _, err := client.Rest().SetGuildCommands(client.ApplicationID(), config.Config.Discord.GuildId, application_commands); err != nil {
		slog.Error("error setting guild commands",
			slog.Any("error", err),
			slog.Any("application_id", client.ApplicationID()),
			slog.Any("guild_id", config.Config.Discord.GuildId),
		)
		os.Exit(1)
		return
	}
	slog.Info("set guild commands", slog.Int("command_count", len(application_commands)))

	if err := client.OpenGateway(context.Background()); err != nil {
		slog.Error("error opening gateway", slog.Any("error", err))
		os.Exit(1)
		return
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	sig := <-s

	slog.Info("shutting down", slog.String("signal", sig.String()))
	client.SetPresence(context.Background(), gateway.WithOnlineStatus(discord.OnlineStatusInvisible))
	slog.Info("closing client")
	client.Close(context.Background())
	slog.Info("goodbye")
	os.Exit(0)
}
