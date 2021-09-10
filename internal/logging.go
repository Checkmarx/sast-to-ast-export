package internal

import (
	"io"

	"github.com/rs/zerolog"
)

type MultiLevelWriter struct {
	io.Writer
	isVerbose       bool
	consoleMinLevel zerolog.Level
	consoleWriter   io.Writer
	fileWriter      io.Writer
}

func NewMultiLevelWriter(isVerbose bool, consoleMinLevel zerolog.Level, consoleWriter, fileWriter io.Writer) MultiLevelWriter {
	return MultiLevelWriter{
		isVerbose:       isVerbose,
		consoleMinLevel: consoleMinLevel,
		consoleWriter:   consoleWriter,
		fileWriter:      fileWriter,
	}
}

// WriteLevel writes to all applicable writers, given the verbosity level and logging level
// if one writer returns an error, the writing stops and the error is returned
// the number of bytes writen is the sum of bytes written to all writers
func (mlw *MultiLevelWriter) WriteLevel(l zerolog.Level, p []byte) (int, error) {
	var total, n int
	var err error
	for _, w := range mlw.getWriters(l) {
		n, err = w.Write(p)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, err
}

// getWriters returns the appropriate writers depending on current log level and verbosity
func (mlw *MultiLevelWriter) getWriters(l zerolog.Level) []io.Writer {
	if !mlw.isVerbose && l < mlw.consoleMinLevel {
		return []io.Writer{mlw.fileWriter}
	}
	return []io.Writer{mlw.consoleWriter, mlw.fileWriter}
}
