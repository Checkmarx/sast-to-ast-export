package internal

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
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

	filePath := path.Join(e.TmpDir, fileName)
	return ioutil.WriteFile(filePath, data, FilePerm)
}

func (e *Export) CreateExportPackage(prefix, outputPath string) (string, error) {
	tmpZipFile, err := ioutil.TempFile(e.TmpDir, fmt.Sprintf("%s.*.zip", prefix))
	if err != nil {
		return "", err
	}

	chdirErr := os.Chdir(e.TmpDir)
	if chdirErr != nil {
		return "", chdirErr
	}

	zipErr := CreateZipFile(tmpZipFile, e.FileList)
	if zipErr != nil {
		return "", zipErr
	}
	tmpZipFileName := tmpZipFile.Name()

	// encrypt zip and key
	zipContents, err := ioutil.ReadFile(tmpZipFileName)
	if err != nil {
		return "", err
	}

	symmetricKey, keyErr := CreateSymmetricKey(SymmetricKeySize)
	if keyErr != nil {
		return "", keyErr
	}

	zipCiphertext, aesErr := AESEncrypt(symmetricKey, zipContents)
	if aesErr != nil {
		return "", aesErr
	}

	symmetricKeyCiphertext, rsaErr := RSAEncrypt([]byte(RSAPublicKey), symmetricKey)
	if rsaErr != nil {
		return "", rsaErr
	}

	// write encrypted zip and key to files
	if ioErr := ioutil.WriteFile(EncryptedKeyFileName, symmetricKeyCiphertext, FilePerm); ioErr != nil {
		return "", ioErr
	}
	if ioErr := ioutil.WriteFile(EncryptedZipFileName, zipCiphertext, FilePerm); ioErr != nil {
		return "", ioErr
	}

	// create final zip with encrypted files
	exportFileName := path.Join(outputPath, CreateExportFileName(prefix, time.Now()))
	exportFile, ioErr := os.Create(exportFileName)
	if ioErr != nil {
		return "", ioErr
	}
	defer exportFile.Close()

	exportErr := CreateZipFile(exportFile, []string{EncryptedKeyFileName, EncryptedZipFileName})
	return exportFileName, exportErr
}

func (e *Export) Clean() error {
	return os.RemoveAll(e.TmpDir)
}

func CreateExportFileName(prefix string, now time.Time) string {
	return fmt.Sprintf("%s-%s.zip", prefix, now.Format("2006-01-02-15-04-05"))
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
