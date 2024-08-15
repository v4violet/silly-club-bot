package main

import (
	"testing"

	"go.uber.org/fx"
)

func TestApp(m *testing.T) {
	if err := fx.ValidateApp(app...); err != nil {
		m.Fatal(err)
	}
}
