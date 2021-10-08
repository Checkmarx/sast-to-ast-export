package export

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateExport(t *testing.T) {
	prefix := "cxsast-create-export"
	export, err := CreateExport(prefix)
	assert.NoError(t, err)
	defer func() {
		closeErr := export.Clean()
		assert.NoError(t, closeErr)
	}()

	info, statErr := os.Stat(export.TmpDir)
	assert.NoError(t, statErr)
	assert.True(t, info.IsDir())
	assert.Contains(t, export.TmpDir, prefix)
}

func TestExport_GetTmpDir(t *testing.T) {
	prefix := "cxsast-get-tmp-dir"
	export, err := CreateExport(prefix)
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
	prefix := "cxsast-add-file-with-data-source"
	t.Run("success case", func(t *testing.T) {
		export, err := CreateExport(prefix)
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
		assert.Equal(t, expectedFileList, export.FileList)

		test1FileName := path.Join(export.TmpDir, "test1.txt")
		info, statErr := os.Stat(test1FileName)
		assert.NoError(t, statErr)
		assert.False(t, info.IsDir())

		content, ioErr := ioutil.ReadFile(test1FileName)
		assert.NoError(t, ioErr)
		assert.Equal(t, "this is test1", string(content))
	})
	t.Run("fails if data source fails", func(t *testing.T) {
		export, err := CreateExport(prefix)
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
	prefix := "cxsast-create-export-package"
	t.Run("success case", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir(os.TempDir(), prefix)
		assert.NoError(t, err)
		defer func(path string) {
			removeErr := os.RemoveAll(path)
			assert.NoError(t, removeErr)
		}(tmpDir)

		export, err := CreateExport(prefix)
		assert.NoError(t, err)
		defer func(export *Export) {
			cleanErr := export.Clean()
			assert.NoError(t, cleanErr)

			_, statErr := os.Stat(export.TmpDir)
			assert.Error(t, statErr)
			assert.True(t, os.IsNotExist(statErr))
		}(&export)

		addErr1 := export.AddFile("test1.txt", []byte("this is test1"))
		assert.NoError(t, addErr1)

		addErr2 := export.AddFile("test2.txt", []byte("this is test2"))
		assert.NoError(t, addErr2)

		exportFileName, exportErr := export.CreateExportPackage(prefix, tmpDir)
		assert.NoError(t, exportErr)

		info, statErr := os.Stat(exportFileName)
		assert.NoError(t, statErr)
		assert.False(t, info.IsDir())
		assert.Contains(t, exportFileName, prefix)

		zipReader, zipErr := zip.OpenReader(exportFileName)
		assert.NoError(t, zipErr)
		defer func(zipReader *zip.ReadCloser) {
			closeErr := zipReader.Close()
			assert.NoError(t, closeErr)
		}(zipReader)

		encryptedKeyFile, zipErr := zipReader.Open(EncryptedKeyFileName)
		assert.NoError(t, zipErr)

		_, keyStatErr := encryptedKeyFile.Stat()
		assert.NoError(t, keyStatErr)

		encryptedZipFile, zipErr := zipReader.Open(EncryptedZipFileName)
		assert.NoError(t, zipErr)

		_, zipStatErr := encryptedZipFile.Stat()
		assert.NoError(t, zipStatErr)
	})
	t.Run("fails if tmp folder doesn't exist", func(t *testing.T) {
		tmpDir, err := ioutil.TempDir(os.TempDir(), prefix)
		assert.NoError(t, err)
		defer func(path string) {
			removeErr := os.RemoveAll(path)
			assert.NoError(t, removeErr)
		}(tmpDir)

		export, err := CreateExport(prefix)
		assert.NoError(t, err)

		cleanErr := export.Clean()
		assert.NoError(t, cleanErr)

		exportFileName, exportErr := export.CreateExportPackage(prefix, tmpDir)

		assert.Error(t, exportErr)
		assert.Equal(t, "", exportFileName)
	})
	t.Run("fails if output folder doesn't exist", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), prefix, "does", "not", "exist")

		export, err := CreateExport(prefix)
		assert.NoError(t, err)
		defer func(export *Export) {
			cleanErr := export.Clean()
			assert.NoError(t, cleanErr)

			_, statErr := os.Stat(export.TmpDir)
			assert.Error(t, statErr)
			assert.True(t, os.IsNotExist(statErr))
		}(&export)

		addErr1 := export.AddFile("test1.txt", []byte("this is test1"))
		assert.NoError(t, addErr1)

		addErr2 := export.AddFile("test2.txt", []byte("this is test2"))
		assert.NoError(t, addErr2)

		exportFileName, exportErr := export.CreateExportPackage(prefix, tmpDir)

		assert.Error(t, exportErr)
		assert.Equal(t, "", exportFileName)
	})
}

func TestExport_Clean(t *testing.T) {
	prefix := "cxsast-clean"
	export, err := CreateExport(prefix)
	assert.NoError(t, err)

	cleanErr := export.Clean()
	assert.NoError(t, cleanErr)

	_, statErr := os.Stat(export.TmpDir)
	assert.Error(t, statErr)
	assert.True(t, os.IsNotExist(statErr))
}

func TestCreateExportFileName(t *testing.T) {
	prefix := "cxsast-create-export-file-name"
	now := time.Date(2021, time.August, 18, 12, 27, 34, 0, time.UTC)

	result := CreateExportFileName(prefix, now)

	expected := fmt.Sprintf("%s-2021-08-18-12-27-34.zip", prefix)
	assert.Equal(t, expected, result)
}

func TestCreateZipFile(t *testing.T) {
	prefix := "cxsast-create-zip-file"
	t.Run("success case", func(t *testing.T) {
		tmpDir, tmpDirErr := ioutil.TempDir(os.TempDir(), prefix)
		assert.NoError(t, tmpDirErr)
		defer func(path string) {
			removeErr := os.RemoveAll(path)
			assert.NoError(t, removeErr)
		}(tmpDir)

		zipFileName := filepath.Join(tmpDir, "test.zip")
		zipFile, zipErr := os.Create(zipFileName)
		assert.NoError(t, zipErr)

		defer func() {
			closeErr := zipFile.Close()
			assert.NoError(t, closeErr)
		}()

		test1FileName := filepath.Join(tmpDir, "test1.txt")
		test1File, test1Err := os.Create(test1FileName)
		assert.NoError(t, test1Err)

		test1CloseErr := test1File.Close()
		assert.NoError(t, test1CloseErr)

		err := CreateZipFile(zipFile, []string{test1FileName})
		assert.NoError(t, err)
	})
	t.Run("fails if zip file doesn't exist", func(t *testing.T) {
		tmpDir, tmpDirErr := ioutil.TempDir(os.TempDir(), prefix)
		assert.NoError(t, tmpDirErr)
		defer func(path string) {
			removeErr := os.RemoveAll(path)
			assert.NoError(t, removeErr)
		}(tmpDir)

		test1FileName := filepath.Join(tmpDir, "test1.txt")
		test1File, test1Err := os.Create(test1FileName)
		assert.NoError(t, test1Err)

		test1CloseErr := test1File.Close()
		assert.NoError(t, test1CloseErr)

		err := CreateZipFile(nil, []string{test1FileName})
		assert.Error(t, err)
	})
	t.Run("fails if zipped file doesn't exist", func(t *testing.T) {
		tmpDir, tmpDirErr := ioutil.TempDir(os.TempDir(), prefix)
		assert.NoError(t, tmpDirErr)
		defer func(path string) {
			removeErr := os.RemoveAll(path)
			assert.NoError(t, removeErr)
		}(tmpDir)

		zipFileName := filepath.Join(tmpDir, "test.zip")
		zipFile, zipErr := os.Create(zipFileName)
		assert.NoError(t, zipErr)

		defer func() {
			closeErr := zipFile.Close()
			assert.NoError(t, closeErr)
		}()

		test1FileName := filepath.Join(tmpDir, "test1.txt")
		test1File, test1Err := os.Create(test1FileName)
		assert.NoError(t, test1Err)

		test1CloseErr := test1File.Close()
		assert.NoError(t, test1CloseErr)

		err := CreateZipFile(zipFile, []string{test1FileName, "doesnt-exist.txt"})
		assert.Error(t, err)
	})
}
