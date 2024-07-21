package main

import (
	"context"

	stdlog "log"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/disgoorg/snowflake/v2"
	_ "go.uber.org/automaxprocs"
)

var buildinfo *debug.BuildInfo
var revision *string

func init() {
	logger := log.New(os.Stderr)
	logger.SetReportTimestamp(true)
	logger.SetReportCaller(true)
	slog.SetDefault(slog.New(logger))

	buildinf, ok := debug.ReadBuildInfo()
	if !ok {
		slog.Warn("unable to read build info")
	} else {
		buildinfo = buildinf
	}
	if val, ok := os.LookupEnv("SOURCE_COMMIT"); ok {
		revision = &val
	}
	for _, kv := range buildinfo.Settings {
		if kv.Key == "vcs.revision" {
			revision = &kv.Value
			break
		}
	}
}

func main() {
	intents := gateway.IntentGuilds | gateway.IntentGuildMembers
	caches := cache.FlagGuilds | cache.FlagMembers

	if isModuleEnabled(ModuleAutoReact) {
		intents |= gateway.IntentGuildMessages | gateway.IntentMessageContent
	}
	if isModuleEnabled(ModuleRandomReact) {
		intents |= gateway.IntentGuildMessages
	}
	if isModuleEnabled(ModuleVoiceLog) {
		intents |= gateway.IntentGuildVoiceStates
		caches |= cache.FlagVoiceStates
	}
	if isModuleEnabled(ModuleVotePin) {
		intents |= gateway.IntentGuildMessageReactions
	}

	slog.Info("starting",
		slog.Any("enabled_modules", dynamicConfig.EnabledModulesRaw),
		slog.Any("intents", intents),
		slog.Any("enabled_caches", caches),
	)

	gatewayConfig := []gateway.ConfigOpt{
		gateway.WithIntents(intents),
		gateway.WithAutoReconnect(true),
		gateway.WithCompress(true),
	}
	if revision != nil {
		gatewayConfig = append(gatewayConfig, gateway.WithPresenceOpts(gateway.WithCustomActivity((*revision)[:7])))
	}
	client, err := disgo.New(dynamicConfig.Discord.Token,
		bot.WithShardManagerConfigOpts(
			sharding.WithShardCount(1),
			sharding.WithAutoScaling(true),
			sharding.WithGatewayConfigOpts(gatewayConfig...),
		),
		bot.WithEventManagerConfigOpts(
			bot.WithListenerFunc(onMessageCreate),
			bot.WithListenerFunc(onGuildMemberJoin),
			bot.WithListenerFunc(onGuildMemberLeave),
			bot.WithListenerFunc(onGuildVoiceStateUpdate),
			bot.WithListenerFunc(onMessageReactionAdd),
			bot.WithListenerFunc(onApplicationCommandInteractionCreate),
			bot.WithListenerFunc(func(event *events.Ready) {
				slog.Info("shard ready",
					slog.Int("shard_id", event.ShardID()),
					slog.String("user_id", event.User.ID.String()),
					slog.String("user_tag", event.User.Tag()),
				)
			}),
			bot.WithListenerFunc(func(event *events.GuildsReady) {
				slog.Info("guilds ready",
					slog.Int("count", event.Client().Caches().GuildsLen()),
					slog.Int("shard_id", event.ShardID()),
				)
			}),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(caches),
		),
	)
	if err != nil {
		stdlog.Fatal(err)
	}

	registerCommands(client)

	if err = client.OpenShardManager(context.Background()); err != nil {
		stdlog.Fatal(err)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	sig := <-s
	slog.Info("shutting down", slog.String("signal", sig.String()))
	for shardId := range client.ShardManager().Shards() {
		client.SetPresenceForShard(context.Background(), shardId, gateway.WithOnlineStatus(discord.OnlineStatusInvisible))
	}
	slog.Info("closing client")
	client.Close(context.Background())
	slog.Info("goodbye")
	os.Exit(0)
}

func registerCommands(client bot.Client) {
	commands := make([]discord.ApplicationCommandCreate, 0, 1)
	if isModuleEnabled(ModuleSetColor) {
		commands = append(commands, discord.SlashCommandCreate{
			Name:        "setcolor",
			Description: "set your custom role color",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "color",
					Description: "hex color",
					Required:    true,
				},
			},
		})
	}
	if _, err := client.Rest().SetGuildCommands(client.ApplicationID(), snowflake.MustParse(dynamicConfig.Discord.GuildId), commands); err != nil {
		slog.Error("error setting guild commands",
			slog.Any("error", err),
			slog.Any("application_id", client.ApplicationID()),
			slog.Any("guild_id", dynamicConfig.Discord.GuildId),
		)
	}
	slog.Info("successfully set guild commands", slog.Int("command_count", len(commands)))
}
