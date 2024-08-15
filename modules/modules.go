package modules

import (
	"text/template"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/v4violet/silly-club-bot/botmodule"
	"go.uber.org/fx"
)

var modules = []fx.Option{}

type Results struct {
	fx.Out

	Options []bot.ConfigOpt `group:"bot,flatten"`
}

type ResultsWithApplicationCommands struct {
	fx.Out

	Options            []bot.ConfigOpt                    `group:"bot,flatten"`
	ApplicationCommand []discord.ApplicationCommandCreate `group:"bot,flatten"`
}

type ApplicationCommandsResults struct {
	fx.Out

	ApplicationCommand []discord.ApplicationCommandCreate `group:"bot,flatten"`
}

type Params struct {
	fx.In

	Client        bot.Client
	DiscordConfig botmodule.DiscordConfig
}

type ParamsWithConfig[T any] struct {
	fx.In

	Client        bot.Client
	DiscordConfig botmodule.DiscordConfig
	Config        T
}

type ParamsWithTemplate struct {
	fx.In

	Client        bot.Client
	DiscordConfig botmodule.DiscordConfig
	Template      *template.Template
}

type ParamsWithConfigAndTemplate[T any] struct {
	fx.In

	Client        bot.Client
	DiscordConfig botmodule.DiscordConfig
	Config        T
	Template      *template.Template
}

func NewModules() fx.Option {
	return fx.Module("modules", modules...)
}
