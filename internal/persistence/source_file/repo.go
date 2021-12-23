package source_file

import (
	"os"
	"path/filepath"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"

	"github.com/pkg/errors"
)

const (
	metadataFilePerm   = 0600
	metadataFolderPerm = 0700
	filesPerBatch      = 10
)

type Repo struct {
	soapClient soap.Adapter
}

func NewRepo(soapClient soap.Adapter) *Repo {
	return &Repo{soapClient: soapClient}
}

func (e *Repo) DownloadSourceFiles(scanID string, sourceFiles map[string]string) error {
	var batches []Batch
	currentBatch := 0
	for k, v := range sourceFiles {
		if len(batches) < currentBatch+1 {
			batches = append(batches, Batch{LocalFiles: []string{}, RemoteFiles: []string{}})
		}
		if _, statErr := os.Stat(v); errors.Is(statErr, os.ErrNotExist) {
			batches[currentBatch].RemoteFiles = append(batches[currentBatch].RemoteFiles, k)
			batches[currentBatch].LocalFiles = append(batches[currentBatch].LocalFiles, v)
		}
		if len(batches[currentBatch].RemoteFiles) > filesPerBatch {
			currentBatch++
		}
	}
	for _, batch := range batches {
		sourceResponse, sourceErr := e.soapClient.GetSourcesByScanID(scanID, batch.RemoteFiles)
		if sourceErr != nil {
			return errors.Wrap(sourceErr, "could not fetch sources")
		}
		contents := sourceResponse.GetSourcesByScanIDResult.CxWSResponseSourcesContent.CxWSResponseSourceContents
		for i, file := range contents {
			createErr := createFileAndPath(batch.LocalFiles[i], []byte(file.Source), metadataFilePerm, metadataFolderPerm)
			if createErr != nil {
				return errors.Wrap(createErr, "could not create file")
			}
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
