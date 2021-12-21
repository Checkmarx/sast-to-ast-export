package export

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/encryption"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	// UsersFileName name of the file with user info
	UsersFileName = "users.json"
	// RolesFileName name of the file with roles info
	RolesFileName = "roles.json"
	// LdapServersFileName name of the file with ldap servers
	LdapServersFileName = "ldap_servers.json"
	// LdapRoleMappingsFileName name of the file with ldap role mapping
	LdapRoleMappingsFileName = "ldap_role_mappings.json"
	// LdapTeamMappingsFileName name of the file with ldap team mappingj
	LdapTeamMappingsFileName = "ldap_team_mappings.json"
	// SamlIdpFileName name of the file about saml idp
	SamlIdpFileName = "saml_identity_providers.json"
	// SamlRoleMappingsFileName saml roles mapping file
	SamlRoleMappingsFileName = "saml_role_mappings.json"
	// SamlTeamMappingsFileName salm teams mapping file
	SamlTeamMappingsFileName = "saml_team_mappings.json"
	// TeamsFileName teams file
	TeamsFileName        = "teams.json"
	encryptedKeyFileName = "key.enc.bin"
	encryptedZipFileName = "zip.enc.bin"

	// DateTimeFormat the formal to use for DT
	DateTimeFormat = "2006-01-02-15-04-05"

	symmetricKeySize = 32
	filePerm         = 0600
)

type Exporter interface {
	AddFile(fileName string, data []byte) error
	AddFileWithDataSource(fileName string, dataSource func() ([]byte, error)) error
	CreateExportPackage(prefix, outputPath string) (string, error)
	Clean() error
	GetTmpDir() string
}

type Export struct {
	TmpDir   string
	FileList []string
}

// CreateExport creates ExportProducer structure and temporary directory
// The caller is responsible for calling the Export.Clear function
// when it's done with the ExportProducer
func CreateExport(prefix string) (Export, error) {
	tmpDir := os.TempDir()
	tmpExportDir, err := os.MkdirTemp(tmpDir, prefix)
	return Export{TmpDir: tmpExportDir, FileList: []string{}}, err
}

func (e *Export) GetTmpDir() string {
	return e.TmpDir
}

// AddFile creates a file with the specified name and content in
// ExportProducer's temporary directory.
func (e *Export) AddFile(fileName string, data []byte) error {
	e.FileList = append(e.FileList, fileName)

	filePath := path.Join(e.TmpDir, fileName)
	return os.WriteFile(filePath, data, filePerm)
}

// AddFileWithDataSource creates the specified file with content provided by dataSource
func (e *Export) AddFileWithDataSource(fileName string, dataSource func() ([]byte, error)) error {
	content, err := dataSource()
	if err != nil {
		return err
	}
	return e.AddFile(fileName, content)
}

// CreateExportPackage compresses and encrypts all files added so far
// nolint:funlen,gocyclo
func (e *Export) CreateExportPackage(prefix, outputPath string) (string, error) {
	tmpZipFile, err := os.CreateTemp(e.TmpDir, fmt.Sprintf("%s.*.zip", prefix))
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := tmpZipFile.Close(); closeErr != nil {
			log.Debug().Err(closeErr).Msg("closing tmp zip file")
		}
	}()

	initialPath, getwdErr := os.Getwd()
	if getwdErr != nil {
		return "", getwdErr
	}

	if chdirErr := os.Chdir(e.TmpDir); chdirErr != nil {
		return "", chdirErr
	}
	defer func() {
		if chdirErr := os.Chdir(initialPath); chdirErr != nil {
			log.Debug().Err(chdirErr).Msg("changing back to initial path")
		}
	}()

	zipErr := createZipFile(tmpZipFile, e.FileList)
	if zipErr != nil {
		return "", zipErr
	}
	tmpZipFileName := tmpZipFile.Name()

	// encrypt zip and key
	zipTmp, err := os.Open(tmpZipFileName)
	if err != nil {
		return "", errors.Wrap(err, "failed to open tmp zip file")
	}
	defer zipTmp.Close()
	zipOut, err := os.Create(encryptedZipFileName)
	if err != nil {
		return "", errors.Wrap(err, "failed to create enc zip file")
	}
	defer zipOut.Close()

	symmetricKey, keyErr := encryption.CreateSymmetricKey(symmetricKeySize)
	if keyErr != nil {
		return "", errors.Wrap(keyErr, "failed to create key for aes")
	}

	aesErr := encryption.EncryptSymmetric(zipTmp, zipOut, symmetricKey)
	if aesErr != nil {
		return "", errors.Wrap(aesErr, "failed to encrypt data")
	}

	keyBytes, decodeErr := base64.StdEncoding.DecodeString(encryption.BuildTimeRSAPublicKey)
	if decodeErr != nil {
		return "", errors.Wrap(decodeErr, "failed to base64 decode RSA public key")
	}

	publicKey, keyErr := encryption.CreatePublicKeyFromKeyBytes(keyBytes)
	if keyErr != nil {
		return "", errors.Wrap(keyErr, "failed to encrypt key")
	}

	symmetricKeyCiphertext, rsaErr := encryption.EncryptAsymmetric(publicKey, symmetricKey)
	if rsaErr != nil {
		return "", errors.Wrap(rsaErr, "rsa encryption failed on key")
	}

	// write encrypted zip and key to files
	if ioErr := os.WriteFile(encryptedKeyFileName, symmetricKeyCiphertext, filePerm); ioErr != nil {
		return "", errors.Wrap(ioErr, "failed to write key to FS")
	}

	// create final zip with encrypted files
	exportFileName := path.Join(outputPath, CreateExportFileName(prefix, time.Now()))
	exportFile, ioErr := os.Create(exportFileName)
	if ioErr != nil {
		return "", errors.Wrap(ioErr, "failed to create file for encrypted data")
	}
	defer func() {
		if closeErr := exportFile.Close(); closeErr != nil {
			log.Debug().Err(closeErr).Msg("closing export file")
		}
	}()

	exportErr := createZipFile(exportFile, []string{encryptedKeyFileName, encryptedZipFileName})
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

// createZipFile zips the list of files and saves into the specified file handle
func createZipFile(zipFile *os.File, fileList []string) error {
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

	return zipWriter.Close()
}
