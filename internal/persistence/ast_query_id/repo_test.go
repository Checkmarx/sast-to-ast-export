package ast_query_id

import (
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/stretchr/testify/assert"
)

func TestRepo_GetQueryID(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		repo, repoErr := NewRepo(AllQueries)
		assert.NoError(t, repoErr)

		result, err := repo.GetQueryID("Kotlin", "Code_Injection", "Kotlin_High_Risk")
		assert.NoError(t, err)
		assert.Equal(t, "15158446363146771540", result)
	})
	t.Run("fails if source path doesn't exist", func(t *testing.T) {
		repo, repoErr := NewRepo(AllQueries)
		assert.NoError(t, repoErr)

		result, err := repo.GetQueryID("Language", "Name", "Group")
		assert.EqualError(t, err, "unknown source path: queries/Language/Group/Name/Name.cs")
		assert.Equal(t, "", result)
	})
}

func TestRepo_GetAllQueryIDsByGroup(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		repo, repoErr := NewRepo(AllQueries)
		assert.NoError(t, repoErr)

		result, err := repo.GetAllQueryIDsByGroup("Kotlin", "Code_Injection")
		assert.NoError(t, err)
		expected := []interfaces.ASTQuery{
			{Language: "Kotlin", Name: "Code_Injection", Group: "Kotlin_High_Risk", QueryID: "15158446363146771540"},
		}
		assert.Equal(t, expected, result)
	})
	t.Run("fails if source path doesn't exist", func(t *testing.T) {
		repo, repoErr := NewRepo(AllQueries)
		assert.NoError(t, repoErr)

		result, err := repo.GetAllQueryIDsByGroup("Language", "Name")
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})
}
