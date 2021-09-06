package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUtils_dataToJSONReader(t *testing.T) {
	reportBody := &ReportRequest{
		ReportType: ReportType,
		ScanID:     1000000,
	}

	t.Run("convert data successfully", func(t *testing.T) {

		result := dataToJSONReader(reportBody)
		assert.NotNil(t, result)
		assert.Equal(t, result, "{\"reportType\":\"XML\",\"scanId\":1000000}")
	})

	t.Run("convert data unsuccessfully", func(t *testing.T) {

		result := dataToJSONReader(nil)
		assert.Nil(t, result)
	})
}
