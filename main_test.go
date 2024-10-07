package main

import (
	"testing"

	"github.com/v4violet/silly-club-bot/botmodule"
	"github.com/v4violet/silly-club-bot/modules"
	"go.uber.org/fx"
)

func TestApp(m *testing.T) {
	if err := modules.ProcessAutoReactions(); err != nil {
		m.Fatal(err)
	}
	app = append(app, fx.Supply(botmodule.DryRun{DryRun: true}))
	if err := fx.ValidateApp(app...); err != nil {
		m.Fatal(err)
	}
}
