package main

import (
	"testing"

	"github.com/v4violet/silly-club-bot/botmodule"
	"go.uber.org/fx"
)

func TestApp(m *testing.T) {
	app = append(app, fx.Supply(botmodule.DryRun{DryRun: true}))
	if err := fx.ValidateApp(app...); err != nil {
		m.Fatal(err)
	}
}
