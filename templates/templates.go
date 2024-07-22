package templates

import (
	"bytes"
	"embed"
	"log/slog"
	"os"
	"text/template"
)

//go:embed *
var templatesFs embed.FS

var Template *template.Template

func Init() {
	slog.Info("loading templates")
	tmpl, err := template.ParseFS(templatesFs, "*.tmpl", "*/*.tmpl")
	if err != nil {
		slog.Error("error parsing templates", slog.Any("error", err))
		os.Exit(1)
		return
	}
	Template = tmpl
	slog.Info("loaded templates")
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
