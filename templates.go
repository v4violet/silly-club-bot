package main

import (
	"bytes"
	"embed"
	"log"
	"log/slog"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS
var templates *template.Template

func init() {
	tmpl, err := template.ParseFS(templateFS, "templates/*.tmpl", "templates/*/*.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	templates = tmpl
	slog.Info("loaded templates")
}

func t(ident string, data any) string {
	var out bytes.Buffer
	err := templates.ExecuteTemplate(&out, ident, data)
	if err != nil {
		panic(err)
	}
	outStr := out.String()
	return outStr
}
