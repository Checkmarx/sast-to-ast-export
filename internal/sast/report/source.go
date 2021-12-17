package report

import (
	"fmt"
	"path/filepath"

	"github.com/checkmarxDev/ast-sast-export/internal/database/store"
	"github.com/pkg/errors"
)

type SourceProvider interface {
	GetBasePath(scanID string) (string, error)
}

type Source struct {
	configStore    store.CxComponentConfigurationStore
	taskScansStore store.TaskScansStore
}

func NewSource(configStore store.CxComponentConfigurationStore, taskScansStore store.TaskScansStore) *Source {
	return &Source{
		configStore:    configStore,
		taskScansStore: taskScansStore,
	}
}

func (e *Source) GetBasePath(scanID string) (string, error) {
	sourcePathConfig, sourcePathErr := e.configStore.GetByKey("SOURCE_PATH")
	if sourcePathErr != nil {
		return "", errors.Wrap(sourcePathErr, "could not fetch source path")
	}
	taskScan, taskScanErr := e.taskScansStore.GetByID(scanID)
	if taskScanErr != nil {
		return "", errors.Wrap(sourcePathErr, "could not fetch task scan")
	}
	basePath := filepath.Join(sourcePathConfig.Value, fmt.Sprintf("%d_%s", taskScan.ProjectID, taskScan.SourceID))
	return basePath, nil
}
