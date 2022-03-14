package sourcefile

import (
	"os"
	"path/filepath"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"

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

func (e *Repo) DownloadSourceFiles(scanID string, sourceFiles []interfaces.SourceFile) error {
	var batches []Batch
	currentBatch := 0
	for _, v := range sourceFiles {
		if len(batches) < currentBatch+1 {
			batches = append(batches, Batch{LocalFiles: []string{}, RemoteFiles: []string{}})
		}
		if _, statErr := os.Stat(v.LocalName); errors.Is(statErr, os.ErrNotExist) {
			batches[currentBatch].RemoteFiles = append(batches[currentBatch].RemoteFiles, v.RemoteName)
			batches[currentBatch].LocalFiles = append(batches[currentBatch].LocalFiles, v.LocalName)
		}
		if len(batches[currentBatch].RemoteFiles) >= filesPerBatch {
			currentBatch++
		}
	}
	for _, batch := range batches {
		sourceResponse, sourceErr := e.soapClient.GetSourcesByScanID(scanID, batch.RemoteFiles)
		if sourceErr != nil {
			return errors.Wrapf(sourceErr, "could not fetch sources scanID=%s remoteFiles=%v", scanID, batch.RemoteFiles)
		}
		contents := sourceResponse.GetSourcesByScanIDResult.CxWSResponseSourcesContent.CxWSResponseSourceContents
		for i, file := range contents {
			createErr := createFileAndPath(batch.LocalFiles[i], []byte(file.Source), metadataFilePerm, metadataFolderPerm)
			if createErr != nil {
				return errors.Wrapf(createErr, "could not create local file scanID=%s localFile=%s", scanID, batch.LocalFiles[i])
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
