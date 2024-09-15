package main

import (
	"flag"
	"os"

	"github.com/v4violet/silly-club-bot/botmodule"
	"github.com/v4violet/silly-club-bot/emojis"
	"github.com/v4violet/silly-club-bot/logmodule"
	"github.com/v4violet/silly-club-bot/modules"
	"github.com/v4violet/silly-club-bot/templates"
	"go.uber.org/fx"
)

var app = []fx.Option{
	logmodule.Module,
	templates.Module,
	emojis.Module,
	botmodule.Module,
	modules.NewModules(),
}

func main() {
	var fxlog = false
	flag.BoolVar(&fxlog, "fxlog", false, "")
	var fxgraph = ""
	flag.StringVar(&fxgraph, "fxgraph", "", "")
	flag.Parse()
	if !fxlog {
		app = append(app, fx.NopLogger)
	}
	if len(fxgraph) > 0 {
		app = append(app, fx.Supply(botmodule.DryRun{true}), fx.Invoke(func(graph fx.DotGraph, shutdowner fx.Shutdowner) error {
			f, err := os.Create(fxgraph)
			if err != nil {
				return err
			}
			if _, err := f.Write([]byte(graph)); err != nil {
				return err
			}
			shutdowner.Shutdown()
			return nil
		}))
	} else {
		app = append(app, fx.Supply(botmodule.DryRun{false}))
	}
	fx.New(app...).Run()
}
