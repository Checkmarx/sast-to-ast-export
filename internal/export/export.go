package export

import (
	"archive/zip"
	"compress/flate"
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
	TeamsFileName = "teams.json"

	encryptedKeyFileName = "key.enc.bin"

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
	tmpDir   string
	fileList []string
}

// CreateExport creates ExportProducer structure and temporary directory
// The caller is responsible for calling the Export.Clear function
// when it's done with the ExportProducer
func CreateExport(prefix string) (Export, error) {
	tmpDir := os.TempDir()
	tmpExportDir, err := os.MkdirTemp(tmpDir, prefix)
	return Export{tmpDir: tmpExportDir, fileList: []string{}}, err
}

func (e *Export) GetTmpDir() string {
	return e.tmpDir
}

// AddFile creates a file with the specified name and content in
// ExportProducer's temporary directory.
func (e *Export) AddFile(fileName string, data []byte) error {
	e.fileList = append(e.fileList, fileName)

	filePath := path.Join(e.tmpDir, fileName)
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
func (e *Export) CreateExportPackage(prefix, outputPath string) (string, error) {
	// create zip
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

	// get key
	symmetricKey, keyErr := encryption.CreateSymmetricKey(symmetricKeySize)
	if keyErr != nil {
		return "", errors.Wrap(keyErr, "failed to create key for aes")
	}
	encKey, err := getEncryptedKey(symmetricKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate encrypted key for archive")
	}

	zipWriter := zip.NewWriter(exportFile)
	defer zipWriter.Close()

	zipKeyWriter, err := zipWriter.Create(encryptedKeyFileName)
	if err != nil {
		return "", errors.Wrap(err, "failed to open zip writer for key")
	}
	_, err = zipKeyWriter.Write(encKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to write to zip key file")
	}

	// chain deflate and encryption and then deflate of the zip archive itself
	// the first deflate is needed to reduce the encrypted file size
	// otherwise if file is encrypted and then deflated, the deflate won't be able to reduce the size
	// because of the chaoric bytes of encrypted data
	for _, fileName := range e.fileList {
		err = func() error {
			// files are added to the tmp
			file, ferr := os.Open(path.Join(e.tmpDir, fileName))
			if ferr != nil {
				return errors.Wrap(ferr, "failed to open file for zip")
			}
			defer file.Close()

			zipFileWriter, zerr := zipWriter.Create(fileName)
			if zerr != nil {
				return errors.Wrapf(zerr, "failed to open zip writer for file %s", fileName)
			}

			// create pipe (bytes written to pw go to pr)
			pr, pw := io.Pipe()
			// errChan needs to be buffered to not block the pipe
			errChan := make(chan error, 1)
			go func() {
				// operations with pipe writer need to be in a separate goroutine
				// writer needs to be closed for reader to stop "waiting" for new bytes
				defer pw.Close()
				// apply first DEFLATE to original content (which will come to pipe writer from file)
				// this will send DEFLATEd content down the pipe to the reader
				flateWriter, ferr := flate.NewWriter(pw, flate.DefaultCompression)
				if ferr != nil {
					errChan <- err
					return
				}
				defer flateWriter.Close()
				if _, err = io.Copy(flateWriter, file); err != nil {
					errChan <- err
					return
				}
				errChan <- nil
			}()
			// EncryptSymmetric will get the DEFLATEd content from pipe reader, encrypt it and send
			// to zipFileWriter, which will apply DEFLATE again and write bytes inside the zip archive
			err = encryption.EncryptSymmetric(pr, zipFileWriter, symmetricKey)
			if err != nil {
				return errors.Wrap(err, "failed to encrypt data")
			}

			return <-errChan
		}()
		if err != nil {
			return "", err
		}
	}

	return exportFileName, nil
}

// Clean removes ExportProducer's temporary directory and it's contents
func (e *Export) Clean() error {
	return os.RemoveAll(e.tmpDir)
}

// CreateExportFileName creates a file name with the format: {prefix}-yyyy-mm-dd-HH-MM-SS.zip
func CreateExportFileName(prefix string, now time.Time) string {
	return fmt.Sprintf("%s-%s.zip", prefix, now.Format(DateTimeFormat))
}

func getEncryptedKey(key []byte) (encKey []byte, err error) {
	keyBytes, err := base64.StdEncoding.DecodeString(encryption.BuildTimeRSAPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to base64 decode RSA public key")
	}

	publicKey, err := encryption.CreatePublicKeyFromKeyBytes(keyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt key")
	}

	encKey, err = encryption.EncryptAsymmetric(publicKey, key)
	return
}
