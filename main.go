package main

import (
	"github.com/v4violet/silly-club-bot/botmodule"
	"github.com/v4violet/silly-club-bot/emojis"
	"github.com/v4violet/silly-club-bot/logmodule"
	"github.com/v4violet/silly-club-bot/modules"
	"github.com/v4violet/silly-club-bot/templates"
	"go.uber.org/fx"
)

var app = []fx.Option{
	fx.NopLogger,
	logmodule.Module,
	templates.Module,
	emojis.Module,
	botmodule.Module,
	modules.NewModules(),
}

func main() {
	fx.New(app...).Run()
}
