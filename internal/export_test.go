package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateExportFileName(t *testing.T) {
	now := time.Date(2021, time.August, 18, 12, 27, 34, 0, time.UTC)
	prefix := "test"

	result := CreateExportFileName(prefix, now)

	expected := "test-2021-08-18-12-27-34.zip"
	assert.Equal(t, expected, result)
}
