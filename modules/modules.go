package modules

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
)

type Module struct {
	Name                string
	Init                func() []bot.ConfigOpt
	ApplicationCommands *[]discord.ApplicationCommandCreate
}

var Modules = make(map[string]Module)

func RegisterModule(module Module) {
	if module.Name == "all" {
		slog.Warn("module should not be named `all`")
	}
	if _, duplicate := Modules[module.Name]; duplicate {
		slog.Warn("duplicate module registered", slog.String("module", module.Name))
	}
	Modules[module.Name] = module
	slog.Info("registered module", slog.String("module", module.Name))

}
