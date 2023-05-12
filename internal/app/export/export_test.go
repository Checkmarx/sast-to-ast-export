package export

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/app/encryption"
	"github.com/stretchr/testify/assert"
)

func TestCreateExport(t *testing.T) {
	prefix := "cxsast-test-create-export"
	export, err := CreateExport(prefix, time.Now())
	assert.NoError(t, err)
	defer func() {
		closeErr := export.Clean()
		assert.NoError(t, closeErr)
	}()

	info, statErr := os.Stat(export.tmpDir)
	assert.NoError(t, statErr)
	assert.True(t, info.IsDir())
	assert.Contains(t, export.tmpDir, prefix)
}

func TestCreateExportLocal(t *testing.T) {
	prefix := "cxsast-test-create-export-local"
	export, err := CreateExportLocal(prefix, time.Now())
	assert.NoError(t, err)
	defer func() {
		closeErr := export.Clean()
		assert.NoError(t, closeErr)
	}()

	info, statErr := os.Stat(export.tmpDir)
	assert.NoError(t, statErr)
	assert.True(t, info.IsDir())
	assert.Contains(t, export.tmpDir, prefix)
}

func TestExport_GetTmpDir(t *testing.T) {
	prefix := "cxsast-test-export-get-tmp-dir"
	export, err := CreateExport(prefix, time.Now())
	assert.NoError(t, err)
	defer func() {
		closeErr := export.Clean()
		assert.NoError(t, closeErr)
	}()

	result := export.GetTmpDir()

	assert.DirExists(t, result)
	assert.Contains(t, result, prefix)
}

func TestExport_AddFileWithDataSource(t *testing.T) {
	prefix := "cxsast-test-export-add-file-with-data-source"
	runTime := time.Now()

	t.Run("success case", func(t *testing.T) {
		export, err := CreateExport(prefix, runTime)
		assert.NoError(t, err)
		defer func() {
			closeErr := export.Clean()
			assert.NoError(t, closeErr)
		}()
		dataSource := func() ([]byte, error) {
			return []byte("this is test1"), nil
		}
		addErr := export.AddFileWithDataSource("test1.txt", dataSource)
		assert.NoError(t, addErr)

		expectedFileList := []string{"test1.txt"}
		assert.Equal(t, expectedFileList, export.fileList)

		test1FileName := path.Join(export.tmpDir, "test1.txt")
		info, statErr := os.Stat(test1FileName)
		assert.NoError(t, statErr)
		assert.False(t, info.IsDir())

		content, ioErr := os.ReadFile(test1FileName)
		assert.NoError(t, ioErr)
		assert.Equal(t, "this is test1", string(content))
	})
	t.Run("fails if data source fails", func(t *testing.T) {
		export, err := CreateExport(prefix, runTime)
		assert.NoError(t, err)
		defer func() {
			closeErr := export.Clean()
			assert.NoError(t, closeErr)
		}()
		dataSource := func() ([]byte, error) {
			return []byte{}, fmt.Errorf("data source error")
		}
		addErr := export.AddFileWithDataSource("test1.txt", dataSource)
		assert.EqualError(t, addErr, "data source error")
	})
}
func TestExportLocal_AddFileWithDataSource(t *testing.T) {
	prefix := "cxsast-test-export-local-add-file-with-data-source"
	runTime := time.Now()

	t.Run("success case", func(t *testing.T) {
		export, err := CreateExportLocal(prefix, runTime)
		assert.NoError(t, err)
		defer func() {
			closeErr := export.Clean()
			assert.NoError(t, closeErr)
		}()
		dataSource := func() ([]byte, error) {
			return []byte("this is test1"), nil
		}
		addErr := export.AddFileWithDataSource("test1.txt", dataSource)
		assert.NoError(t, addErr)

		expectedFileList := []string{"test1.txt"}
		assert.Equal(t, expectedFileList, export.fileList)

		test1FileName := path.Join(export.tmpDir, "test1.txt")
		info, statErr := os.Stat(test1FileName)
		assert.NoError(t, statErr)
		assert.False(t, info.IsDir())

		content, ioErr := os.ReadFile(test1FileName)
		assert.NoError(t, ioErr)
		assert.Equal(t, "this is test1", string(content))
	})
	t.Run("fails if data source fails", func(t *testing.T) {
		export, err := CreateExportLocal(prefix, runTime)
		assert.NoError(t, err)
		defer func() {
			closeErr := export.Clean()
			assert.NoError(t, closeErr)
		}()
		dataSource := func() ([]byte, error) {
			return []byte{}, fmt.Errorf("data source error")
		}
		addErr := export.AddFileWithDataSource("test1.txt", dataSource)
		assert.EqualError(t, addErr, "data source error")
	})
}

func TestExport_CreateExportPackage(t *testing.T) {
	prefix := "cxsast-test-export-create-export-package"
	runTime := time.Now()

	t.Run("success case", func(t *testing.T) {
		tmpDir := createTmpDir(t, prefix)
		defer clearTmpDir(t, tmpDir)

		export, err := CreateExport(prefix, runTime)
		assert.NoError(t, err)
		defer func(export *Export) {
			cleanErr := export.Clean()
			assert.NoError(t, cleanErr)

			_, statErr := os.Stat(export.tmpDir)
			assert.Error(t, statErr)
			assert.True(t, os.IsNotExist(statErr))
		}(&export)

		files := map[string][]byte{
			"test1.txt": []byte("this is test1"),
			"test2.txt": []byte("this is test2"),
		}
		for fname, content := range files {
			addErr := export.AddFile(fname, content)
			assert.NoError(t, addErr)
		}

		exportFileName, keyFileName, exportErr := export.CreateExportPackage(prefix, tmpDir)
		assert.NoError(t, exportErr)

		info, statErr := os.Stat(exportFileName)
		assert.NoError(t, statErr)
		assert.False(t, info.IsDir())
		assert.Contains(t, exportFileName, prefix)

		keyInfo, keyStatErr := os.Stat(keyFileName)
		assert.NoError(t, keyStatErr)
		assert.False(t, keyInfo.IsDir())
		assert.Contains(t, keyFileName, prefix)

		encodedKey, keyIOErr := os.ReadFile(keyFileName)
		assert.NoError(t, keyIOErr)

		key, base64Err := base64.StdEncoding.DecodeString(string(encodedKey))
		assert.NoError(t, base64Err)

		zipReader, zipErr := zip.OpenReader(exportFileName)
		assert.NoError(t, zipErr)
		defer func(zipReader *zip.ReadCloser) {
			closeErr := zipReader.Close()
			assert.NoError(t, closeErr)
		}(zipReader)

		// test that zip has files and they are encrypted
		for fname, content := range files {
			zr, zipFileErr := zipReader.Open(fname)
			assert.NoError(t, zipFileErr)
			bt, zipFileIOErr := io.ReadAll(zr)
			assert.NoError(t, zipFileIOErr)
			assert.NotEqual(t, content, bt)
			// decrypt zipped content
			encryptedFile := bytes.NewBuffer(bt)
			compressedFile := bytes.NewBuffer([]byte{})
			decryptErr := encryption.DecryptSymmetric(encryptedFile, compressedFile, key)
			assert.NoError(t, decryptErr)
			// decompress decrypted content
			flateReader := flate.NewReader(compressedFile)
			plaintext, flateErr := io.ReadAll(flateReader)
			assert.NoError(t, flateErr)
			assert.Equal(t, content, plaintext)
		}
	})
	t.Run("fails if output folder doesn't exist", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), prefix, "does", "not", "exist")

		export, err := CreateExport(prefix, runTime)
		assert.NoError(t, err)
		defer func(export *Export) {
			cleanErr := export.Clean()
			assert.NoError(t, cleanErr)

			_, statErr := os.Stat(export.tmpDir)
			assert.Error(t, statErr)
			assert.True(t, os.IsNotExist(statErr))
		}(&export)

		addErr1 := export.AddFile("test1.txt", []byte("this is test1"))
		assert.NoError(t, addErr1)

		addErr2 := export.AddFile("test2.txt", []byte("this is test2"))
		assert.NoError(t, addErr2)

		exportFileName, _, exportErr := export.CreateExportPackage(prefix, tmpDir)

		assert.Error(t, exportErr)
		assert.Equal(t, "", exportFileName)
	})
}

func TestExport_Clean(t *testing.T) {
	prefix := "cxsast-test-export-clean"
	export, err := CreateExport(prefix, time.Now())
	assert.NoError(t, err)

	cleanErr := export.Clean()
	assert.NoError(t, cleanErr)

	_, statErr := os.Stat(export.tmpDir)
	assert.Error(t, statErr)
	assert.True(t, os.IsNotExist(statErr))
}

func TestCreateExportFileName(t *testing.T) {
	now := time.Date(2021, time.August, 18, 12, 27, 34, 0, time.UTC)
	tests := []struct {
		Prefix,
		Suffix,
		Extension string
		Now      time.Time
		Expected string
	}{
		{"cxsast-create-export-file-name", "", "zip", now, "cxsast-create-export-file-name-2021-08-18-12-27-34.zip"},
		{"prefix", "suffix", "txt", now, "prefix-2021-08-18-12-27-34-suffix.txt"},
	}

	for i, test := range tests {
		d := test
		t.Run(fmt.Sprintf("#%d", i+1), func(t *testing.T) {
			result := CreateExportFileName(d.Prefix, d.Suffix, d.Extension, d.Now)

			assert.Equal(t, d.Expected, result)
		})
	}
}

func TestCreateDir(t *testing.T) {
	prefix := "cxsast-test-create-dir"
	export, err := CreateExport(prefix, time.Now())
	assert.NoError(t, err)
	defer func() {
		closeErr := export.Clean()
		assert.NoError(t, closeErr)
	}()

	testDirName := "test_name"
	errDir := export.CreateDir(testDirName)

	assert.NoError(t, errDir)
	tempDirName := export.GetTmpDir()

	assert.DirExists(t, path.Join(tempDirName, testDirName))
}

func TestWriteSymmetricKeyToFile(t *testing.T) {
	prefix := "cxsast-write-symmetric-key-to-file"

	t.Run("success case", func(t *testing.T) {
		tmpDir := createTmpDir(t, prefix)
		defer clearTmpDir(t, tmpDir)
		keyFilename := path.Join(tmpDir, "key.txt")
		key := []byte("test")

		err := writeSymmetricKeyToFile(keyFilename, key)

		assert.NoError(t, err)
		data, ioErr := os.ReadFile(keyFilename)
		assert.NoError(t, ioErr)
		expected := base64.StdEncoding.EncodeToString(key)
		assert.Equal(t, expected, string(data))
	})
	t.Run("fails if can't write to file", func(t *testing.T) {
		tmpDir := createTmpDir(t, prefix)
		defer clearTmpDir(t, tmpDir)
		keyFilename := path.Join(tmpDir, "invalid", "key.txt")
		key := []byte("test")

		err := writeSymmetricKeyToFile(keyFilename, key)

		assert.Error(t, err)
	})
}

func createTmpDir(t *testing.T, prefix string) string {
	tmpDir, tmpDirErr := os.MkdirTemp(os.TempDir(), prefix)
	assert.NoError(t, tmpDirErr)
	return tmpDir
}

func clearTmpDir(t *testing.T, tmpPath string) {
	removeErr := os.RemoveAll(tmpPath)
	assert.NoError(t, removeErr)
}
