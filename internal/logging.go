package internal

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

var (
	levelMapping = map[int]slog.Level{
		0: slog.LevelError,
		1: slog.LevelWarn,
		2: slog.LevelInfo,
		3: slog.LevelDebug,
	}
)

func InitLogging(level int) {
	logLevel, ok := levelMapping[level]
	if !ok {
		if level > 3 {
			logLevel = slog.LevelDebug
		} else {
			logLevel = slog.LevelInfo
		}
	}

	base := slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      logLevel,
			TimeFormat: time.DateTime,
		}),
	)
	slog.SetDefault(base)
}
