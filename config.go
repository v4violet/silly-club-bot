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

	EnabledModulesRaw []Module `env:"ENABLED_MODULES, default=all"`
	EnabledModules    map[Module]bool
}

type DiscordConfig struct {
	Token   string `env:"TOKEN, required"`
	GuildId string `env:"GUILD_ID"`
}

type StaticConfig struct {
	GuildID            string               `json:"guild_id"`
	RandomReaction     RandomReactionConfig `json:"random_reaction"`
	JoinRoles          JoinRolesConfig      `json:"joinroles"`
	UserLogChannel     string               `json:"user_log_channel"`
	VoiceLimitChannel  string               `json:"voice_limit_channel"`
	VotePinMinVotes    int                  `json:"vote_pin_min_votes"`
	SetColorLogChannel string               `json:"setcolor_log_channel"`

	AutoReactRaw map[string]string `json:"autoreact"`
	AutoReact    map[*regexp.Regexp]string
}

type RandomReactionConfig struct {
	Percentage float32 `json:"percentage"`
	Emoji      string  `json:"emoji"`
	LogChannel string  `json:"log_channel"`

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

	dynamicConfig.EnabledModules = make(map[Module]bool, len(dynamicConfig.EnabledModulesRaw))

	for _, module := range dynamicConfig.EnabledModulesRaw {
		dynamicConfig.EnabledModules[module] = true
	}

	if err := json.Unmarshal([]byte(staticConfigRaw), &staticConfig); err != nil {
		log.Fatal(err)
	}

	staticConfig.AutoReact = make(map[*regexp.Regexp]string, len(staticConfig.AutoReactRaw))

	for regexRaw, emojiId := range staticConfig.AutoReactRaw {
		staticConfig.AutoReact[regexp.MustCompile(regexRaw)] = emojiId
	}

	staticConfig.RandomReaction.ChannelWhitelist = make(map[string]bool, len(staticConfig.RandomReaction.ChannelWhitelistRaw))

	for _, channelId := range staticConfig.RandomReaction.ChannelWhitelistRaw {
		staticConfig.RandomReaction.ChannelWhitelist[channelId] = true
	}

	if len(dynamicConfig.Discord.GuildId) <= 0 {
		dynamicConfig.Discord.GuildId = staticConfig.GuildID
	}
}

type Module string

const (
	ModuleJoinRoles   Module = "join_roles"
	ModuleUserLog     Module = "user_log"
	ModuleVoiceLimit  Module = "voice_limit"
	ModuleAutoReact   Module = "auto_react"
	ModuleRandomReact Module = "random_react"
	ModuleVotePin     Module = "vote_pin"
	ModuleVoiceLog    Module = "voice_log"
	ModuleSetColor    Module = "set_color"
	ModuleAll         Module = "all"
)

func isModuleEnabled(module Module) bool {
	_, all_enabled := dynamicConfig.EnabledModules[ModuleAll]
	_, disabled := dynamicConfig.EnabledModules[module]
	return all_enabled || disabled
}
