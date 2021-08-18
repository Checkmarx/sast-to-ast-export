package internal

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	prefix = "test"
)

func TestCreateExport(t *testing.T) {
	export, err := CreateExport(prefix)
	assert.NoError(t, err)

	defer export.Clean()
	info, statErr := os.Stat(export.TmpDir)
	assert.NoError(t, statErr)
	assert.True(t, info.IsDir())
	assert.Contains(t, export.TmpDir, prefix)
}

func TestExportClean(t *testing.T) {
	export, err := CreateExport(prefix)
	assert.NoError(t, err)

	cleanErr := export.Clean()
	assert.NoError(t, cleanErr)

	_, statErr := os.Stat(export.TmpDir)
	assert.Error(t, statErr)
	assert.True(t, os.IsNotExist(statErr))
}

func TestCreateExportFileName(t *testing.T) {
	now := time.Date(2021, time.August, 18, 12, 27, 34, 0, time.UTC)

	result := CreateExportFileName(prefix, now)

	expected := "test-2021-08-18-12-27-34.zip"
	assert.Equal(t, expected, result)
}
