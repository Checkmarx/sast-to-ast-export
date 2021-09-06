package internal

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"
)

const (
	SymmetricKeySize = 32
	FilePerm         = 0600

	UsersFileName            = "users.json"
	RolesFileName            = "roles.json"
	LdapServersFileName      = "ldap_servers.json"
	LdapRoleMappingsFileName = "ldap_role_mappings.json"
	LdapTeamMappingsFileName = "ldap_team_mappings.json"
	SamlIdpFileName          = "saml_identity_providers.json"
	SamlRoleMappingsFileName = "saml_role_mappings.json"
	SamlTeamMappingsFileName = "saml_team_mappings.json"
	TeamsFileName            = "teams.json"
	EncryptedKeyFileName     = "key.enc.bin"
	EncryptedZipFileName     = "zip.enc.bin"

	DateTimeFormat = "2006-01-02-15-04-05"
)

type Export struct {
	TmpDir   string
	FileList []string
}

// CreateExport creates ExportProducer structure and temporary directory
// The caller is responsible for calling the Export.Clear function
// when it's done with the ExportProducer
func CreateExport(prefix string) (Export, error) {
	tmpDir := os.TempDir()
	tmpExportDir, err := ioutil.TempDir(tmpDir, prefix)
	return Export{TmpDir: tmpExportDir, FileList: []string{}}, err
}

// AddFile creates a file with the specified name and content in
// ExportProducer's temporary directory.
func (e *Export) AddFile(fileName string, data []byte) error {
	e.FileList = append(e.FileList, fileName)

	filePath := path.Join(e.TmpDir, fileName)
	return ioutil.WriteFile(filePath, data, FilePerm)
}

// CreateExportPackage compresses and encrypts all files added so far
func (e *Export) CreateExportPackage(prefix, outputPath string) (string, error) {
	tmpZipFile, err := ioutil.TempFile(e.TmpDir, fmt.Sprintf("%s.*.zip", prefix))
	if err != nil {
		return "", err
	}
	defer tmpZipFile.Close()

	initialPath, getwdErr := os.Getwd()
	if getwdErr != nil {
		return "", getwdErr
	}

	if chdirErr := os.Chdir(e.TmpDir); chdirErr != nil {
		return "", chdirErr
	}
	defer os.Chdir(initialPath)

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

	zipCiphertext, aesErr := EncryptSymmetric(symmetricKey, zipContents)
	if aesErr != nil {
		return "", aesErr
	}

	keyBytes, decodeErr := base64.StdEncoding.DecodeString(buildTimeRSAPublicKey)
	if decodeErr != nil {
		return "", decodeErr
	}

	publicKey, keyErr := CreatePublicKeyFromKeyBytes(keyBytes)
	if keyErr != nil {
		return "", keyErr
	}

	symmetricKeyCiphertext, rsaErr := EncryptAsymmetric(publicKey, symmetricKey)
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

// Clean removes ExportProducer's temporary directory and it's contents
func (e *Export) Clean() error {
	return os.RemoveAll(e.TmpDir)
}

// CreateExportFileName creates a file name with the format: {prefix}-yyyy-mm-dd-HH-MM-SS.zip
func CreateExportFileName(prefix string, now time.Time) string {
	return fmt.Sprintf("%s-%s.zip", prefix, now.Format(DateTimeFormat))
}

// CreateZipFile zips the list of files and saves into the specified file handle
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
