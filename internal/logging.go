package internal

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Logger zerolog.Logger
)

func InitLogging(level string) {
	parsedLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid log level")
	}

	Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.DateTime}).Level(parsedLevel)
}
