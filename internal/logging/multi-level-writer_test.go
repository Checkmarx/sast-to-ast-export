package logging

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type mockWriter struct {
	WriteHandler func(p []byte) (n int, err error)
}

func (e *mockWriter) Write(p []byte) (n int, err error) {
	return e.WriteHandler(p)
}

func TestMultiLevelWriter_WriteLevel(t *testing.T) {
	t.Run("writes debug to file only if not verbose", func(t *testing.T) {
		mockConsoleWriter := new(bytes.Buffer)
		mockFileWriter := new(bytes.Buffer)
		writer := NewMultiLevelWriter(false, zerolog.InfoLevel, mockConsoleWriter, mockFileWriter)

		result, err := writer.WriteLevel(zerolog.DebugLevel, []byte("test"))

		assert.NoError(t, err)
		assert.Equal(t, 4, result)
		assert.Equal(t, "", mockConsoleWriter.String())
		assert.Equal(t, "test", mockFileWriter.String())
	})
	t.Run("writes debug to file and console if verbose", func(t *testing.T) {
		mockConsoleWriter := new(bytes.Buffer)
		mockFileWriter := new(bytes.Buffer)
		writer := NewMultiLevelWriter(true, zerolog.InfoLevel, mockConsoleWriter, mockFileWriter)

		result, err := writer.WriteLevel(zerolog.DebugLevel, []byte("test"))

		assert.NoError(t, err)
		assert.Equal(t, 8, result)
		assert.Equal(t, "test", mockConsoleWriter.String())
		assert.Equal(t, "test", mockFileWriter.String())
	})
	t.Run("writes info to file and console if not verbose", func(t *testing.T) {
		mockConsoleWriter := new(bytes.Buffer)
		mockFileWriter := new(bytes.Buffer)
		writer := NewMultiLevelWriter(false, zerolog.InfoLevel, mockConsoleWriter, mockFileWriter)

		result, err := writer.WriteLevel(zerolog.InfoLevel, []byte("test"))

		assert.NoError(t, err)
		assert.Equal(t, 8, result)
		assert.Equal(t, "test", mockConsoleWriter.String())
		assert.Equal(t, "test", mockFileWriter.String())
	})
	t.Run("fails without writing, if console writer fails to write", func(t *testing.T) {
		mockConsoleWriter := &mockWriter{WriteHandler: func(p []byte) (n int, err error) {
			return 0, fmt.Errorf("write error")
		}}
		mockFileWriter := new(bytes.Buffer)
		writer := NewMultiLevelWriter(false, zerolog.InfoLevel, mockConsoleWriter, mockFileWriter)

		result, err := writer.WriteLevel(zerolog.InfoLevel, []byte("test"))

		assert.EqualError(t, err, "write error")
		assert.Equal(t, 0, result)
		assert.Equal(t, "", mockFileWriter.String())
	})
	t.Run("fails with writing, if file writer fails to write", func(t *testing.T) {
		mockConsoleWriter := new(bytes.Buffer)
		mockFileWriter := &mockWriter{WriteHandler: func(p []byte) (n int, err error) {
			return 0, fmt.Errorf("write error")
		}}
		writer := NewMultiLevelWriter(false, zerolog.InfoLevel, mockConsoleWriter, mockFileWriter)

		result, err := writer.WriteLevel(zerolog.InfoLevel, []byte("test"))

		assert.EqualError(t, err, "write error")
		assert.Equal(t, 4, result)
		assert.Equal(t, "test", mockConsoleWriter.String())
	})
}
