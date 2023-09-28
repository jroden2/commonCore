package utils

import (
	"github.com/rs/zerolog"
	"io"
	"os"
)

func InitialiseLogger(Output ...io.Writer) zerolog.Logger {
	if Output != nil {
		logger := zerolog.New(Output[0]).With().Timestamp().Logger()
		return logger
	}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	return logger
}

func LogError(functionName string, err error, logger zerolog.Logger) {
	logger.Error().Stack().Caller(1).Str("function", functionName).Err(err).Send()
}

func LogInfo(functionName, message string, logger zerolog.Logger) {
	logger.Info().Stack().Caller(1).Str("function", functionName).Str("log", message).Send()
}

func LogFatal(functionName string, err error, logger zerolog.Logger) {
	logger.Fatal().Stack().Caller(1).Str("function", functionName).Err(err).Send()
}
