package export

import (
	"archive/zip"
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/app/encryption"
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
	// ProjectsFileName projects file
	ProjectsFileName = "projects.json"
	// QueriesFileName queries file
	QueriesFileName = "queries.xml"
	// PresetsDirName presets directory name
	PresetsDirName = "presets"
	// PresetsFileName presets file
	PresetsFileName = "presets.json"
	// InstallationFileName file
	InstallationFileName = "installation.json"

	// DateTimeFormat the formal to use for DT
	DateTimeFormat = "2006-01-02-15-04-05"

	symmetricKeySize = 32
	filePerm         = 0600
	dirPerm          = 0750
)

type Exporter interface {
	AddFile(fileName string, data []byte) error
	AddFileWithDataSource(fileName string, dataSource func() ([]byte, error)) error
	CreateExportPackage(prefix, outputPath string) (string, string, error)
	Clean() error
	GetTmpDir() string
	CreateDir(dirName string) error
}

type Export struct {
	tmpDir   string
	fileList []string
	runTime  time.Time
}

// CreateExport creates ExportProducer structure and temporary directory
// The caller is responsible for calling the Export.Clear function
// when it's done with the ExportProducer
func CreateExport(prefix string, runTime time.Time) (Export, error) {
	tmpDir := os.TempDir()
	tmpExportDir, err := os.MkdirTemp(tmpDir, prefix)
	return Export{tmpDir: tmpExportDir, fileList: []string{}, runTime: runTime}, err
}

// CreateExportLocal creates ExportProducer structure and specified local directory
// The caller is responsible for calling the Export.Clear function
// when it's done with the ExportProducer
func CreateExportLocal(outputPath string, runTime time.Time) (Export, error) {
	err := os.Mkdir(outputPath, dirPerm)
	return Export{tmpDir: outputPath, fileList: []string{}, runTime: runTime}, err
}

// CreateExportFromLocal creates ExportProducer structure using an existing local
// directory and adds all existing files from the local directory into the
// ExportProducer's fileList
// The caller is responsible for calling the Export.Clear function
// when it's done with the ExportProducer
func CreateExportFromLocal(inputPath string, runTime time.Time) (Export, error) {
	fileList := []string{}
	_, err := os.Stat(inputPath)
	if !os.IsNotExist(err) {
		err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return err
			}

			if !info.IsDir() {
				fileList = append(fileList, filepath.ToSlash(strings.TrimPrefix(path, inputPath+string(os.PathSeparator))))
			}

			return nil
		})
	}
	return Export{tmpDir: inputPath, fileList: fileList, runTime: runTime}, err
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
func (e *Export) CreateExportPackage(prefix, outputPath string) (string, string, error) { //nolint:gocritic
	keyFileName := path.Join(outputPath, CreateExportFileName(prefix, "key", "txt", e.runTime))
	exportFileName := path.Join(outputPath, CreateExportFileName(prefix, "", "zip", e.runTime))
	// create zip
	exportFile, ioErr := os.Create(exportFileName)
	if ioErr != nil {
		return "", "", errors.Wrap(ioErr, "failed to create file for encrypted data")
	}
	defer func() {
		if closeErr := exportFile.Close(); closeErr != nil {
			log.Debug().Err(closeErr).Msg("closing export file")
		}
	}()

	// get key
	symmetricKey, keyErr := encryption.CreateSymmetricKey(symmetricKeySize)
	if keyErr != nil {
		return "", "", errors.Wrap(keyErr, "failed to create key for aes")
	}

	zipWriter := zip.NewWriter(exportFile)
	defer zipWriter.Close()

	keyWriteErr := writeSymmetricKeyToFile(keyFileName, symmetricKey)
	if keyWriteErr != nil {
		return "", "", errors.Wrap(keyWriteErr, "failed to write symmetric key to file")
	}

	// chain deflate and encryption and then deflate of the zip archive itself
	// the first deflate is needed to reduce the encrypted file size
	// otherwise if file is encrypted and then deflated, deflate won't be able to reduce the size
	// because of the chaoric bytes of encrypted data
	var err error
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
			return "", "", err
		}
	}

	return exportFileName, keyFileName, nil
}

// Clean removes ExportProducer's temporary directory and it's contents
func (e *Export) Clean() error {
	return os.RemoveAll(e.tmpDir)
}

// CreateExportFileName creates a file name with the format: {prefix}-yyyy-mm-dd-HH-MM-SS.{extension}
func CreateExportFileName(prefix, suffix, extension string, now time.Time) string {
	if suffix != "" {
		return fmt.Sprintf("%s-%s-%s.%s", prefix, now.Format(DateTimeFormat), suffix, extension)
	}
	return fmt.Sprintf("%s-%s.%s", prefix, now.Format(DateTimeFormat), extension)
}

// CreateDir creates directory inside temp directory
func (e *Export) CreateDir(dirName string) error {
	fullDirName := path.Join(e.tmpDir, dirName)
	return os.Mkdir(fullDirName, dirPerm)
}

func NewJSONDataSource(obj interface{}) func() ([]byte, error) {
	return func() ([]byte, error) {
		return json.Marshal(obj)
	}
}

func writeSymmetricKeyToFile(fileName string, key []byte) error {
	encodedKey := base64.StdEncoding.EncodeToString(key)
	return os.WriteFile(fileName, []byte(encodedKey), filePerm)
}
