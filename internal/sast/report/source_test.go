package report

import (
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/database"
	mock_store "github.com/checkmarxDev/ast-sast-export/test/mocks/database/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSource_GetBasePath(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		taskID := "100023"
		config := database.ComponentConfiguration{Value: "C:\\path\\to\\source"}
		taskScan := database.TaskScan{ID: 100023, ProjectID: 4, SourceID: "0000_1111_2222"}
		ctrl := gomock.NewController(t)
		componentsMock := mock_store.NewMockCxComponentConfigurationStore(ctrl)
		componentsMock.EXPECT().GetByKey(gomock.Any()).Return(&config, nil)
		taskScansMock := mock_store.NewMockTaskScansStore(ctrl)
		taskScansMock.EXPECT().GetByID(taskID).Return(&taskScan, nil)
		i := NewSource(componentsMock, taskScansMock)

		result, err := i.GetBasePath(taskID)
		assert.NoError(t, err)
		assert.Equal(t, "C:\\path\\to\\source\\4_0000_1111_2222", result)
	})
}
