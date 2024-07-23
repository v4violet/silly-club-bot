package config

import (
	"context"
	"log/slog"

	"github.com/charmbracelet/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

var Config struct {
	Discord struct {
		Token    string        `env:"TOKEN,required"`
		GuildId  snowflake.ID  `env:"GUILD_ID,required"`
		AuthorId *snowflake.ID `env:"AUTHOR_ID,noinit"`
	} `env:",prefix=DISCORD_"`

	Modules struct {
		EnabledRaw  []string `env:"ENABLED,default=all"`
		DisabledRaw []string `env:"DISABLED"`

		Enabled  map[string]bool
		Disabled map[string]bool

		UserLog struct {
			ChannelId *snowflake.ID `env:"CHANNEL,noinit"`
		} `env:",prefix=USER_LOG_"`

		RandomReact struct {
			LogChannelId             *snowflake.ID  `env:"LOG_CHANNEL,noinit"`
			WhitelistedChannelIdsRaw []snowflake.ID `env:"WHITELISTED_CHANNELS"`
			WhitelistedChannelIds    map[snowflake.ID]bool
		} `env:",prefix=RANDOM_REACT_"`

		JoinRoles struct {
			UserRoleIds []snowflake.ID `env:"USER_ROLES"`
			BotRoleIds  []snowflake.ID `env:"BOT_ROLES"`
		} `env:",prefix=JOIN_ROLES_"`

		SetColor struct {
			LogChannelId *snowflake.ID `env:"LOG_CHANNEL,noinit"`
		} `env:",prefix=SET_COLOR_"`

		VoiceLimit struct {
			ChannelId *snowflake.ID `env:"CHANNEL,noinit"`
		} `env:",prefix=VOICE_LIMIT_"`
	} `env:",prefix=MODULES_"`

	LogLevelRaw string `env:"LOG_LEVEL,default=INFO"`
	LogLevel    slog.Level

	GitRevision *string `env:"SOURCE_COMMIT,noinit"`
}

func Init() {
	godotenv.Load(".env")
	godotenv.Load(".env.default")

	snowflake.AllowUnquoted = true
	if err := envconfig.Process(context.Background(), &Config); err != nil {
		log.Fatal(err)
	}
	snowflake.AllowUnquoted = false

	Config.LogLevel.UnmarshalText([]byte(Config.LogLevelRaw))

	Config.Modules.Enabled = make(map[string]bool)
	Config.Modules.Disabled = make(map[string]bool)

	for _, module := range Config.Modules.EnabledRaw {
		Config.Modules.Enabled[module] = true
	}
	for _, module := range Config.Modules.DisabledRaw {
		Config.Modules.Disabled[module] = true
	}

	Config.Modules.RandomReact.WhitelistedChannelIds = make(map[snowflake.ID]bool, len(Config.Modules.RandomReact.WhitelistedChannelIdsRaw))

	for _, channelId := range Config.Modules.RandomReact.WhitelistedChannelIdsRaw {
		Config.Modules.RandomReact.WhitelistedChannelIds[channelId] = true
	}
}

func ModuleEnabled(module string) bool {
	_, allEnabled := Config.Modules.Enabled["all"]
	_, enabled := Config.Modules.Enabled[module]
	_, disabled := Config.Modules.Disabled[module]
	if disabled {
		return false
	}
	return allEnabled || enabled
}
