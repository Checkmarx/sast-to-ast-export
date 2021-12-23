package metadata

import (
	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	mock_integration_similarity "github.com/checkmarxDev/ast-sast-export/test/mocks/integration/similarity"
	mock_integration_soap "github.com/checkmarxDev/ast-sast-export/test/mocks/integration/soap"
	mock_persistence_ast_query_id "github.com/checkmarxDev/ast-sast-export/test/mocks/persistence/ast_query_id"
	mock_persistence_source "github.com/checkmarxDev/ast-sast-export/test/mocks/persistence/source"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMetadataFactory_GetMetadataForQueryAndResult(t *testing.T) {
	astQueryID := "12532796926860742976"
	firstMethodLine := "100"
	lastMethodLine := "101"
	similarityID := "1234567890"
	scanID := "1000001"
	metaQuery := &MetadataQuery{
		QueryID:  "6300",
		Language: "Kotlin",
		Name:     "SQL_Injection",
		Group:    "Kotlin_High_Risk",
	}
	metaResult := &MetadataResult{
		PathID:   "2",
		ResultID: "1000002",
		FirstNode: MetadataNode{
			FileName: "Goatlin-develop/packages/clients/android/app/src/main/java/com/cx/goatlin/EditNoteActivity.kt",
			Name:     "text",
			Line:     "83",
			Column:   "78",
		},
		LastNode: MetadataNode{
			FileName: "Goatlin-develop/packages/clients/android/app/src/main/java/com/cx/goatlin/helpers/DatabaseHelper.kt",
			Name:     "note",
			Line:     "129",
			Column:   "28",
		},
	}

	ctrl := gomock.NewController(t)
	tmpDir := t.TempDir()
	astQueryIDProviderMock := mock_persistence_ast_query_id.NewMockQueryIDProvider(ctrl)
	astQueryIDProviderMock.EXPECT().GetQueryID(metaQuery.Language, metaQuery.Name, metaQuery.Group).Return(astQueryID, nil)
	similarityIDProviderMock := mock_integration_similarity.NewMockSimilarityIDProvider(ctrl)
	similarityIDProviderMock.EXPECT().Calculate(
		gomock.Any(), metaResult.FirstNode.Name, metaResult.FirstNode.Line, metaResult.FirstNode.Column, firstMethodLine,
		gomock.Any(), metaResult.LastNode.Name, metaResult.LastNode.Line, metaResult.LastNode.Column, lastMethodLine,
		astQueryID,
	).Return(similarityID, nil)
	soapAdapterMock := mock_integration_soap.NewMockAdapter(ctrl)
	soapAdapterMock.EXPECT().GetResultPathsForQuery(scanID, metaQuery.QueryID).Return(&soap.GetResultPathsForQueryResponse{
		GetResultPathsForQueryResult: soap.GetResultPathsForQueryResult{
			Paths: soap.Paths{
				Paths: []soap.ResultPath{
					{
						PathID: metaResult.PathID,
						Node: soap.Node{
							Nodes: []soap.ResultPathNode{
								{MethodLine: firstMethodLine},
								{MethodLine: "2"},
								{MethodLine: "3"},
								{MethodLine: lastMethodLine},
							},
						},
					},
					{
						PathID: "3",
						Node: soap.Node{
							Nodes: []soap.ResultPathNode{{MethodLine: "10"}, {MethodLine: "20"}, {MethodLine: "30"}},
						},
					},
				},
			},
		},
	}, nil)
	sourceProviderMock := mock_persistence_source.NewMockSourceProvider(ctrl)
	sourceProviderMock.EXPECT().
		DownloadSourceFiles(scanID, gomock.Any()).
		DoAndReturn(
			func(_ string, files map[string]string) error {
				expectedFiles := []string{metaResult.FirstNode.FileName, metaResult.LastNode.FileName}
				var result []string
				for k := range files {
					result = append(result, k)
				}
				assert.ElementsMatch(t, expectedFiles, result)
				return nil
			},
		)
	metadata := NewMetadataFactory(astQueryIDProviderMock, similarityIDProviderMock, soapAdapterMock, sourceProviderMock, tmpDir)

	result, err := metadata.GetMetadataForQueryAndResult(scanID, metaQuery, metaResult)
	assert.NoError(t, err)

	expectedResult := MetadataRecord{
		QueryID:      metaQuery.QueryID,
		SimilarityID: similarityID,
		PathID:       metaResult.PathID,
		ResultID:     metaResult.ResultID,
	}
	assert.Equal(t, expectedResult, *result)
}
