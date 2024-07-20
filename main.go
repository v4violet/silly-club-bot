package main

import (
	"context"

	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	_ "go.uber.org/automaxprocs"
)

var buildinfo *debug.BuildInfo
var revision *string

func init() {
	buildinf, ok := debug.ReadBuildInfo()
	if !ok {
		slog.Warn("unable to read build info")
	} else {
		buildinfo = buildinf
	}
	for _, kv := range buildinfo.Settings {
		if kv.Key == "vcs.revision" {
			revision = &kv.Value
			break
		}
	}
}

func main() {
	gatewayConfig := []gateway.ConfigOpt{
		gateway.WithIntents(
			gateway.IntentGuilds,
			gateway.IntentGuildMessages,
			gateway.IntentMessageContent,
			gateway.IntentGuildMembers,
		),
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
			cache.WithCaches(cache.FlagGuilds, cache.FlagChannels, cache.FlagRoles, cache.FlagMembers),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err = client.OpenShardManager(context.Background()); err != nil {
		log.Fatal(err)
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
