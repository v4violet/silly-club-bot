package config

import (
	"context"
	"log"

	"github.com/disgoorg/snowflake/v2"
	"github.com/sethvargo/go-envconfig"
)

var Config struct {
	Discord DiscordConfig `env:",prefix=DISCORD_"`
}

type DiscordConfig struct {
	Token   string       `env:"TOKEN,required"`
	GuildId snowflake.ID `env:"GUILD_ID,required"`
}

func init() {
	snowflake.AllowUnquoted = true
	if err := envconfig.Process(context.Background(), &Config); err != nil {
		log.Fatal(err)
	}
}
