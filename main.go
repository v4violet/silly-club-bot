package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/v4violet/silly-club-bot/config"
	"github.com/v4violet/silly-club-bot/modules"
	"github.com/v4violet/silly-club-bot/modules/auto_react"
	"github.com/v4violet/silly-club-bot/modules/join_roles"
	"github.com/v4violet/silly-club-bot/modules/ping"
	"github.com/v4violet/silly-club-bot/modules/poll_pin"
	"github.com/v4violet/silly-club-bot/modules/random_react"
	"github.com/v4violet/silly-club-bot/modules/set_color"
	"github.com/v4violet/silly-club-bot/modules/user_log"
	"github.com/v4violet/silly-club-bot/modules/voice_limit"
	"github.com/v4violet/silly-club-bot/modules/voice_log"
	"github.com/v4violet/silly-club-bot/modules/vote_pin"
	"github.com/v4violet/silly-club-bot/templates"
	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	config.Init()

	logger := log.New(os.Stdout)

	logger.SetLevel(log.Level(config.Config.LogLevel))

	logger.SetReportCaller(true)
	logger.SetReportTimestamp(true)

	slog.SetDefault(slog.New(logger))

	slog.SetLogLoggerLevel(config.Config.LogLevel)

	if _, err := maxprocs.Set(); err != nil {
		slog.Warn("failed to set GOMAXPROCS", slog.Any("error", err))
	}

	templates.Init()

	auto_react.Init()
	join_roles.Init()
	ping.Init()
	poll_pin.Init()
	random_react.Init()
	set_color.Init()
	user_log.Init()
	voice_limit.Init()
	voice_log.Init()
	vote_pin.Init()
}

func main() {
	slog.Debug("testing build times 0")

	clientConfig := []bot.ConfigOpt{
		bot.WithGatewayConfigOpts(
			gateway.WithCompress(true),
			gateway.WithAutoReconnect(true),
			gateway.WithLogger(slog.Default()),
			gateway.WithIntents(gateway.IntentGuilds),
		),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds)),
		bot.WithEventListenerFunc(func(event *events.Ready) {
			slog.Info("ready",
				slog.String("user_id", event.User.ID.String()),
				slog.String("user_tag", event.User.Tag()),
			)
		}),
		bot.WithEventListenerFunc(func(event *events.GuildsReady) {
			slog.Info("guilds ready", slog.Int("count", event.Client().Caches().GuildsLen()))
		}),
		bot.WithEventListenerFunc(func(event *events.Resumed) {
			slog.Info("resumed", slog.Int("sequence", event.SequenceNumber()))
		}),
	}

	if config.Config.GitRevision != nil {
		clientConfig = append(clientConfig, bot.WithGatewayConfigOpts(gateway.WithPresenceOpts(gateway.WithCustomActivity((*config.Config.GitRevision)[:7]))))
	}

	enabledModules := []string{}

	applicationCommands := []discord.ApplicationCommandCreate{}

	for _, module := range modules.Modules {
		if config.ModuleEnabled(module.Name) {
			enabledModules = append(enabledModules, module.Name)
			clientConfig = append(clientConfig, module.Init()...)
			if module.ApplicationCommands != nil {
				applicationCommands = append(applicationCommands, *module.ApplicationCommands...)
			}
		}
	}

	client, err := disgo.New(config.Config.Discord.Token, clientConfig...)
	if err != nil {
		slog.Error("error constructing bot client", slog.Any("error", err))
		os.Exit(1)
		return
	}

	slog.Info("starting",
		slog.String("enabled_modules", strings.Join(enabledModules, ",")),
		slog.Any("intents", client.Gateway().Intents()),
		slog.Any("caches", client.Caches().CacheFlags()),
		slog.Any("guild_id", config.Config.Discord.GuildId),
	)

	if _, err := client.Rest().SetGuildCommands(client.ApplicationID(), config.Config.Discord.GuildId, applicationCommands); err != nil {
		slog.Error("error setting guild commands",
			slog.Any("error", err),
			slog.Any("application_id", client.ApplicationID()),
			slog.Any("guild_id", config.Config.Discord.GuildId),
		)
		os.Exit(1)
		return
	}
	slog.Info("successfully set guild commands", slog.Int("command_count", len(applicationCommands)))

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
