package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func GetNewConsoleWriter() zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out:             os.Stdout,
		NoColor:         true, // disabled because CMD.exe can't handle escape sequences
		FormatLevel:     func(i interface{}) string { return strings.ToUpper(fmt.Sprintf("%-5s", i)) },
		FormatTimestamp: consoleTimeFormatter,
	}
}

// timeFormatter is heavily inspired on zerolog.consoleDefaultFormatTimestamp
func consoleTimeFormatter(i interface{}) string {
	timeFormat := "15:04:05"
	timeFieldFormat := time.RFC3339
	timeFormatUnixMs := zerolog.TimeFormatUnixMs
	timeFormatUnixMicro := zerolog.TimeFormatUnixMicro
	t := "<nil>"
	switch tt := i.(type) {
	case string:
		ts, err := time.Parse(timeFieldFormat, tt)
		if err != nil {
			t = tt
		} else {
			t = ts.Format(timeFormat)
		}
	case json.Number:
		i, err := tt.Int64()
		if err != nil {
			t = tt.String()
		} else {
			var sec, nsec int64 = i, 0
			switch timeFieldFormat {
			case timeFormatUnixMs:
				nsec = int64(time.Duration(i) * time.Millisecond)
				sec = 0
			case timeFormatUnixMicro:
				nsec = int64(time.Duration(i) * time.Microsecond)
				sec = 0
			}
			ts := time.Unix(sec, nsec).UTC()
			t = ts.Format(timeFormat)
		}
	}
	return t
}
