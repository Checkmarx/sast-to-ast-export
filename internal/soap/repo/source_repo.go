package repo

import (
	"os"
	"path/filepath"

	"github.com/checkmarxDev/ast-sast-export/internal/soap"
	"github.com/pkg/errors"
)

const (
	metadataFilePerm   = 0600
	metadataFolderPerm = 0700
)

type SourceProvider interface {
	DownloadSourceFiles(scanID string, sourceFiles map[string]string) error
}

type SourceRepo struct {
	soapClient soap.Adapter
}

func NewSourceRepo(soapClient soap.Adapter) *SourceRepo {
	return &SourceRepo{soapClient: soapClient}
}

func (e *SourceRepo) DownloadSourceFiles(scanID string, sourceFiles map[string]string) error {
	var remoteFiles []string
	var localFiles []string
	for k, v := range sourceFiles {
		if _, statErr := os.Stat(v); errors.Is(statErr, os.ErrNotExist) {
			remoteFiles = append(remoteFiles, k)
			localFiles = append(localFiles, v)
		}
	}
	sourceResponse, sourceErr := e.soapClient.GetSourcesByScanID(scanID, remoteFiles)
	if sourceErr != nil {
		return errors.Wrap(sourceErr, "could not fetch sources")
	}
	contents := sourceResponse.GetSourcesByScanIDResult.CxWSResponseSourcesContent.CxWSResponseSourceContents
	for i, file := range contents {
		createErr := createFileAndPath(localFiles[i], []byte(file.Source), metadataFilePerm, metadataFolderPerm)
		if createErr != nil {
			return errors.Wrap(createErr, "could not create file")
		}
	}
	return nil
}

func createFileAndPath(filename string, content []byte, filePerm, dirPerm os.FileMode) error {
	pathErr := os.MkdirAll(filepath.Dir(filename), dirPerm)
	if pathErr != nil {
		return pathErr
	}
	return os.WriteFile(filename, content, filePerm)
}
