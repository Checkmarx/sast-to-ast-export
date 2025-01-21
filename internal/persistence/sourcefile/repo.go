package sourcefile

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/rs/zerolog/log"

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

//nolint:gocyclo
func (e *Repo) DownloadSourceFiles(scanID string, sourceFiles []interfaces.SourceFile, rmvdir string) error {
	// Check if rmvdir is provided
	var excludePaths []string
	var excludePatterns []*regexp.Regexp

	if rmvdir != "" {
		var err error
		excludePaths, excludePatterns, err = ReadExcludedPaths(rmvdir)
		if err != nil {
			return err
		}
	}
	// Use the IsExcluded function to check if a file should be excluded
	isExcluded := func(path string) bool {
		for _, exclude := range excludePaths {
			if path == exclude || strings.Contains(path, exclude) {
				return true
			}
		}
		for _, pattern := range excludePatterns {
			if pattern.MatchString(path) {
				return true
			}
		}
		return false
	}

	var batches []Batch
	currentBatch := 0
	for _, v := range sourceFiles {
		if isExcluded(v.RemoteName) {
			log.Info().Msgf("Excluding problematic file: %s", v.RemoteName)
			continue
		}
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
