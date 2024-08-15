package logmodule

import (
	stdlog "log"
	"log/slog"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
	"go.uber.org/fx"
)

var Module = fx.Module("log", fx.Provide(NewCharmLog, NewSlog, NewLog), fx.Invoke(func(log *slog.Logger) {
	slog.SetDefault(log)
}))

func NewCharmLog() *log.Logger {
	log.SetReportTimestamp(true)
	log.SetReportCaller(true)
	log.SetColorProfile(termenv.ANSI256)

	styles := log.DefaultStyles()
	error_style := lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	styles.Keys["err"] = error_style
	styles.Keys["error"] = error_style
	log.SetStyles(styles)

	raw_level, _ := os.LookupEnv("LOG_LEVEL")
	if level, err := log.ParseLevel(raw_level); err == nil {
		log.SetLevel(level)
	}

	return log.Default()
}

func NewSlog(log *log.Logger) *slog.Logger {
	return slog.New(log)
}

func NewLog(log *log.Logger) *stdlog.Logger {
	return log.StandardLog()
}
