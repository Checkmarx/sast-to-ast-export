package internal

import (
	"os"
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

func TestCreateExport(t *testing.T) {
	prefix := "test"
	export, err := CreateExport(prefix)
	assert.NoError(t, err)

	defer export.Clean()
	info, statErr := os.Stat(export.TmpDir)
	assert.NoError(t, statErr)
	assert.True(t, info.IsDir())
	assert.Contains(t, export.TmpDir, prefix)
}

func TestExportClean(t *testing.T) {
	prefix := "test"
	export, err := CreateExport(prefix)
	assert.NoError(t, err)

	cleanErr := export.Clean()
	assert.NoError(t, cleanErr)

	_, statErr := os.Stat(export.TmpDir)
	assert.Error(t, statErr)
	assert.True(t, os.IsNotExist(statErr))
}
