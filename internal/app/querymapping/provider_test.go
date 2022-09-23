package querymapping

import (
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestQueryMappingProvider(t *testing.T) {
	t.Run("Test creating from file", func(t *testing.T) {
		provider, err := NewProvider("../../../data/mapping.json")
		assert.NoError(t, err)

		assert.Equal(t, "../../../data/mapping.json", provider.GetQueryMappingFilePath())
		assert.Equal(t, "11", provider.GetMapping()[0].SastID)
	})

	t.Run("Test creating from URL", func(t *testing.T) {
		provider, err := NewProvider("https://raw.githubusercontent.com/Checkmarx/sast-to-ast-export/feature/AST-16676-add-query-mapping-option/data/mapping.json")
		assert.NoError(t, err)

		var name string
		_, name = path.Split(provider.GetQueryMappingFilePath())

		assert.Equal(t, fileName, name)
		assert.Equal(t, "11", provider.GetMapping()[0].SastID)
		delErr := provider.Clean()
		assert.NoError(t, delErr)
	})

	t.Run("Test error with wrong path", func(t *testing.T) {
		var err error
		_, err = NewProvider("wrong_path")
		assert.Error(t, err)
	})
}
