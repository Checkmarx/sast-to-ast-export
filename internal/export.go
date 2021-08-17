package internal

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	UsersFileName        = "users.json"
	TeamsFileName        = "teams.json"
	EncryptedKeyFileName = "key.enc.bin"
	EncryptedZipFileName = "zip.enc.bin"
	SymmetricKeySize     = 32
	FilePerm             = 0600
)

type Export struct {
	TmpDir   string
	FileList []string
}

func CreateExport(prefix string) (Export, error) {
	tmpDir := os.TempDir()
	tmpExportDir, err := ioutil.TempDir(tmpDir, prefix)
	if err != nil {
		log.Fatal(err)
	}

	return Export{TmpDir: tmpExportDir, FileList: []string{}}, nil
}

func (e *Export) AddFile(fileName string, data []byte) error {
	e.FileList = append(e.FileList, fileName)

	usersFile := path.Join(e.TmpDir, fileName)
	return ioutil.WriteFile(usersFile, data, FilePerm)
}

func (e *Export) CreateZip(prefix string) (string, error) {
	tmpZipFile, err := ioutil.TempFile(e.TmpDir, fmt.Sprintf("%s.*.zip", prefix))
	if err != nil {
		log.Fatal(err)
	}

	zipErr := CreateZipFile(tmpZipFile, e.FileList)

	return tmpZipFile.Name(), zipErr
}

func CreateFileName(basePath, prefix string) string {
	currentTime := time.Now()
	fileName := fmt.Sprintf("%s-%s.zip", prefix, currentTime.Format("2006-01-02-15-04-05"))
	return filepath.Join(basePath, fileName)
}

func CreateZipFile(zipFile *os.File, fileList []string) error {
	zipWriter := zip.NewWriter(zipFile)

	for _, fileName := range fileList {
		// open file to zip
		file, fileErr := os.Open(fileName)
		if fileErr != nil {
			return fileErr
		}

		// create zip entry
		entryFile, zipErr := zipWriter.Create(fileName)
		if zipErr != nil {
			return zipErr
		}

		// copy file to zip entry
		if _, copyErr := io.Copy(entryFile, file); copyErr != nil {
			return copyErr
		}

		// close file
		if closeErr := file.Close(); closeErr != nil {
			return closeErr
		}
	}

	if zipCloseErr := zipWriter.Close(); zipCloseErr != nil {
		return zipCloseErr
	}

	return nil
}
