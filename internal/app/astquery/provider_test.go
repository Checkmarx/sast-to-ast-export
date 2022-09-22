package astquery

import (
	"encoding/xml"
	"fmt"
	"github.com/checkmarxDev/ast-sast-export/internal/app/querymapping"
	"os"
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	mock_interfaces "github.com/checkmarxDev/ast-sast-export/test/mocks/app/ast_query_mapping"
	mock_interfaces_queries "github.com/checkmarxDev/ast-sast-export/test/mocks/app/queries"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type QueryIDTest struct {
	Language, Group, Name, SastID, Expected string
}

type CustomQueriesListTest struct {
	Input, Expected soap.GetQueryCollectionResponse
}

func TestProvider_GetQueryID(t *testing.T) {
	ctrl := gomock.NewController(t)
	queryProvider := mock_interfaces_queries.NewMockQueriesRepo(ctrl)
	queryMappingProvider := mock_interfaces.NewMockQueryMappingRepo(ctrl)

	queryIDTests := []QueryIDTest{
		{"Kotlin", "Kotlin_High_Risk", "Code_Injection", "1", "15158446363146771540"},
		{"CSharp", "General", "Find_SQL_Injection_Evasion_Attack", "2", "8984835614866342550"},
		{"Go", "General", "Find_Command_Injection_Sanitize", "3", "9498204717545098527"},
		{"Go", "Custom", "Custom_query", "100000", "11111"},
	}
	queryMappingProvider.EXPECT().GetMapping().Return([]querymapping.QueryMap{{
		AstID:  "11111",
		SastID: "100000",
	}}).Times(4)
	for _, test := range queryIDTests {
		testName := fmt.Sprintf("%s %s %s", test.Language, test.Group, test.Name)
		t.Run(testName, func(t *testing.T) {
			repo, repoErr := NewProvider(queryProvider, queryMappingProvider)
			assert.NoError(t, repoErr)

			result, err := repo.GetQueryID(test.Language, test.Name, test.Group, test.SastID)
			assert.NoError(t, err)
			assert.Equal(t, test.Expected, result)
		})
	}
}

func TestProvider_GetCustomQueries(t *testing.T) {
	var queriesObj, customQueriesObj soap.GetQueryCollectionResponse
	ctrl := gomock.NewController(t)
	queryProvider := mock_interfaces_queries.NewMockQueriesRepo(ctrl)
	queryMappingProvider := mock_interfaces.NewMockQueryMappingRepo(ctrl)
	queryMappingProvider.EXPECT().GetMapping().Return([]querymapping.QueryMap{{
		AstID:  "11111",
		SastID: "100000",
	}}).Times(1)
	repo, repoErr := NewProvider(queryProvider, queryMappingProvider)
	assert.NoError(t, repoErr)

	t.Run("Successful getting custom queries", func(t *testing.T) {
		queries, ioErr := os.ReadFile("../../../test/data/queries/queries.xml")
		assert.NoError(t, ioErr)
		customQueries, ioCustomErr := os.ReadFile("../../../test/data/queries/custom_queries.xml")
		assert.NoError(t, ioCustomErr)
		_ = xml.Unmarshal(queries, &queriesObj)
		_ = xml.Unmarshal(customQueries, &customQueriesObj)
		queryProvider.EXPECT().GetQueriesList().Return(&queriesObj, nil).Times(1)

		result, err := repo.GetCustomQueriesList()
		assert.NoError(t, err)
		assert.Equal(t, &customQueriesObj, result)
	})

	t.Run("Error with getting custom queries", func(t *testing.T) {
		queryProvider.EXPECT().GetQueriesList().Return(nil, fmt.Errorf("failed getting custom queries")).Times(1)

		_, err := repo.GetCustomQueriesList()
		assert.EqualError(t, err, "failed getting custom queries")
	})

}
