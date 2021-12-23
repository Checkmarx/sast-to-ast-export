package ast_query_id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryIDRepo_GetQueryID(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		repo, repoErr := NewQueryIDRepo(AllQueries)
		assert.NoError(t, repoErr)

		result, err := repo.GetQueryID("Kotlin", "Code_Injection", "Kotlin_High_Risk")
		assert.NoError(t, err)
		assert.Equal(t, "15158446363146771540", result)
	})
	t.Run("fails if source path doesn't exist", func(t *testing.T) {
		repo, repoErr := NewQueryIDRepo(AllQueries)
		assert.NoError(t, repoErr)

		result, err := repo.GetQueryID("Language", "Name", "Group")
		assert.EqualError(t, err, "unknown source path: queries/Language/Group/Name/Name.cs")
		assert.Equal(t, "", result)
	})
}
