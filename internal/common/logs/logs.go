package logs

import (
	"log/slog"
	"os"
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

func mapLogLevel(level string) slog.Level {
	var mappedLevel slog.Level
	err := mappedLevel.UnmarshalText([]byte(level))
	if err != nil {
		mappedLevel = slog.LevelInfo
	}

	return mappedLevel
}
