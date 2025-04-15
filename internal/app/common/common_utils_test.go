package common

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AstQueryIDTest struct {
	Language, Group, Name, Expected string
}

func TestGetAstQueryID(t *testing.T) {
	testRenameFile := filepath.Join("..", "..", "..", "data", "renames.json")

	err := LoadRename(testRenameFile)
	require.NoError(t, err, "Failed to load rename map for test from ../../../data/renames.json")

	astQueryIDTests := []AstQueryIDTest{
		{"Kotlin", "Kotlin_High_Risk", "Code_Injection", "15158446363146771540"},
		{"CSharp", "General", "Find_SQL_Injection_Evasion_Attack", "8984835614866342550"},
		{"Go", "General", "Find_Command_Injection_Sanitize", "9498204717545098527"},
	}

	for _, test := range astQueryIDTests {
		result, err := GetAstQueryID(test.Language, test.Name, test.Group)
		assert.NoError(t, err)
		assert.Equal(t, test.Expected, result)
	}
}
