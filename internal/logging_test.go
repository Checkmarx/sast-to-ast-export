package internal

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestMultiLevelWriter_getWriters(t *testing.T) {
	var mockConsoleWriter, mockFileWriter *bytes.Buffer
	t.Run("returns both writers if not verbose and level equal to info", func(t *testing.T) {
		writer := NewMultiLevelWriter(false, zerolog.InfoLevel, mockConsoleWriter, mockFileWriter)

		result := writer.getWriters(zerolog.InfoLevel)

		assert.Len(t, result, 2)
	})

	t.Run("returns both writers if not verbose and level above info", func(t *testing.T) {
		writer := NewMultiLevelWriter(false, zerolog.InfoLevel, mockConsoleWriter, mockFileWriter)

		result := writer.getWriters(zerolog.WarnLevel)

		assert.Len(t, result, 2)
	})

	t.Run("returns one writer if not verbose and level below info", func(t *testing.T) {
		writer := NewMultiLevelWriter(false, zerolog.InfoLevel, mockConsoleWriter, mockFileWriter)

		result := writer.getWriters(zerolog.DebugLevel)

		assert.Len(t, result, 1)
	})

	t.Run("returns both writers if verbose and level below info", func(t *testing.T) {
		writer := NewMultiLevelWriter(true, zerolog.InfoLevel, mockConsoleWriter, mockFileWriter)

		result := writer.getWriters(zerolog.DebugLevel)

		assert.Len(t, result, 2)
	})
}
