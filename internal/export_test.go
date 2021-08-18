package internal

import (
	"io/ioutil"
	"os"
	"path"
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

func TestExportAddFile(t *testing.T) {
	export, err := CreateExport(prefix)
	assert.NoError(t, err)

	defer export.Clean()
	addErr := export.AddFile("test1.txt", []byte("this is test1"))
	assert.NoError(t, addErr)

	expectedFileList := []string{"test1.txt"}
	assert.Equal(t, expectedFileList, export.FileList)

	test1FileName := path.Join(export.TmpDir, "test1.txt")
	info, statErr := os.Stat(test1FileName)
	assert.NoError(t, statErr)
	assert.False(t, info.IsDir())

	content, ioErr := ioutil.ReadFile(test1FileName)
	assert.NoError(t, ioErr)
	assert.Equal(t, "this is test1", string(content))
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
