//go:build modules.all || modules.auto_react

package modules

import (
	"testing"
)

func TestCompileRegex(m *testing.T) {
	if err := ProcessAutoReactions(); err != nil {
		m.Fatal(err)
	}
}
