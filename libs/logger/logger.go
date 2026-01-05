package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init() {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
    log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func L() zerolog.Logger {
    return log.Logger
}