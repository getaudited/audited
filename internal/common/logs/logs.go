package logs

import (
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

func New(level string) *Logger {
	return &Logger{slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: mapLogLevel(level),
		},
	))}
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}

var levels map[string]slog.Level = map[string]slog.Level{
	"INFO":  slog.LevelInfo,
	"DEBUG": slog.LevelDebug,
}

func mapLogLevel(level string) slog.Level {
	v, exists := levels[strings.ToUpper(level)]
	if !exists {
		return slog.LevelInfo
	}

	return v
}
