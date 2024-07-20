package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"log/slog"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type DynamicConfig struct {
	Discord  *DiscordConfig `env:", prefix=DISCORD_"`
	LogLevel slog.Level     `env:"LOG_LEVEL, default=ERROR"`
}

type DiscordConfig struct {
	Token string `env:"TOKEN, required"`
}

type StaticConfig struct {
	GuildID              string               `json:"guild_id"`
	RandomReaction       RandomReactionConfig `json:"random_reaction"`
	AutoReactUnprocessed map[string]string    `json:"autoreact"`
	AutoReact            map[*regexp.Regexp]string
	JoinRoles            JoinRolesConfig `json:"joinroles"`
	UserLogChannel       string          `json:"user_log_channel"`
	VoiceLimitChannel    string          `json:"voice_limit_channel"`
}

type RandomReactionConfig struct {
	Percentage          float32  `json:"percentage"`
	Emoji               string   `json:"emoji"`
	LogChannel          string   `json:"log_channel"`
	ChannelWhitelistRaw []string `json:"channel_whitelist"`
	ChannelWhitelist    map[string]bool
}

type JoinRolesConfig struct {
	Users []string `json:"users"`
	Bots  []string `json:"bots"`
}

var dynamicConfig DynamicConfig

//go:embed config.json
var staticConfigRaw string
var staticConfig StaticConfig

func init() {
	godotenv.Load()
	if err := envconfig.Process(context.Background(), &dynamicConfig); err != nil {
		log.Fatal(err)
	}

	slog.SetLogLoggerLevel(dynamicConfig.LogLevel)

	if err := json.Unmarshal([]byte(staticConfigRaw), &staticConfig); err != nil {
		log.Fatal(err)
	}

	staticConfig.AutoReact = make(map[*regexp.Regexp]string)

	for regexRaw, emojiId := range staticConfig.AutoReactUnprocessed {
		staticConfig.AutoReact[regexp.MustCompile(regexRaw)] = emojiId
	}

	staticConfig.RandomReaction.ChannelWhitelist = make(map[string]bool)

	for _, channelId := range staticConfig.RandomReaction.ChannelWhitelistRaw {
		staticConfig.RandomReaction.ChannelWhitelist[channelId] = true
	}
}
