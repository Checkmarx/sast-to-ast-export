package astquery

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type QueryIDTest struct {
	Language, Group, Name, Expected string
}

func TestProvider_GetQueryID(t *testing.T) {
	queryIDTests := []QueryIDTest{
		{"Kotlin", "Kotlin_High_Risk", "Code_Injection", "15158446363146771540"},
		{"CSharp", "General", "Find_SQL_Injection_Evasion_Attack", "8984835614866342550"},
		{"Go", "General", "Find_Command_Injection_Sanitize", "9498204717545098527"},
	}
	for _, test := range queryIDTests {
		testName := fmt.Sprintf("%s %s %s", test.Language, test.Group, test.Name)
		t.Run(testName, func(t *testing.T) {
			repo, repoErr := NewProvider()
			assert.NoError(t, repoErr)

			result, err := repo.GetQueryID(test.Language, test.Name, test.Group)
			assert.NoError(t, err)
			assert.Equal(t, test.Expected, result)
		})
	}
}
