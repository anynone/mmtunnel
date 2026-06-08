package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Level string

const (
	Trace   Level = "trace"
	Debug   Level = "debug"
	Info    Level = "info"
	Warning Level = "warning"
	Error   Level = "error"
)

func ParseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(Info):
		return slog.LevelInfo, nil
	case string(Trace):
		return slog.LevelDebug - 4, nil
	case string(Debug):
		return slog.LevelDebug, nil
	case string(Warning), "warn":
		return slog.LevelWarn, nil
	case string(Error):
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported log level %q", value)
	}
}

func New(level string) (*slog.Logger, error) {
	parsed, err := ParseLevel(level)
	if err != nil {
		return nil, err
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: parsed})
	return slog.New(handler), nil
}
