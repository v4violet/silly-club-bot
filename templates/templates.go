package templates

import (
	"bytes"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"text/template"

	"github.com/v4violet/silly-club-bot/emojis"
)

//go:embed *
var templatesFs embed.FS

var Template *template.Template

func init() {
	tmpl, err := template.ParseFS(templatesFs, "*.tmpl", "*/*.tmpl")
	if err != nil {
		slog.Error("error parsing templates", slog.Any("error", err))
		os.Exit(1)
		return
	}

	Template = tmpl
	slog.Info("loaded templates")
}

func LoadEmojis() error {
	for k, v := range emojis.Emojis {
		tmpl, err := Template.New(fmt.Sprintf("emoji.%s", k)).Parse(v.Discord.Mention())
		if err != nil {
			return err
		}
		Template = tmpl
	}
	return nil
}

func Exec(ident string, data any) string {
	var out bytes.Buffer
	err := Template.ExecuteTemplate(&out, ident, data)
	if err != nil {
		panic(err)
	}
	outStr := out.String()
	return outStr
}
