package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

const (
	levelFieldName     string = "level"
	messageFieldName   string = "msg"
	errorFieldName     string = "error"
	timestampFieldName string = "time"
	timeFieldFormat    string = time.RFC3339
)

func Init(logLevel string, outputStream io.Writer) error {
	zerolog.TimeFieldFormat = timeFieldFormat
	zerolog.LevelFieldName = levelFieldName
	zerolog.MessageFieldName = messageFieldName
	zerolog.TimestampFieldName = timestampFieldName
	zerolog.ErrorFieldName = errorFieldName

	err := setLogLevel(logLevel)
	if err != nil {
		return err
	}

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	if outputStream == nil {
		outputStream = os.Stdout
	}

	log.Logger = zerolog.New(outputStream).With().
		Timestamp(). // always add time
		Logger()

	return nil
}

func setLogLevel(logLevel string) error {
	zeroLevel, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(zeroLevel)
	return nil
}
