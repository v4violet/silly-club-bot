package modules

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
)

type Module struct {
	Init                func() ([]bot.ConfigOpt, error)
	ApplicationCommands *[]discord.ApplicationCommandCreate
}

var Modules = map[string]Module{}

func Init() ([]bot.ConfigOpt, error) {
	config := []bot.ConfigOpt{}

	for _, v := range Modules {
		module_config, err := v.Init()
		if err != nil {
			return nil, err
		}
		config = append(config, module_config...)
	}

	return config, nil
}
