package export

import (
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/soap"
	mock_ast "github.com/checkmarxDev/ast-sast-export/test/mocks/ast"
	mock_sast "github.com/checkmarxDev/ast-sast-export/test/mocks/sast"
	mock_soap "github.com/checkmarxDev/ast-sast-export/test/mocks/soap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestMetadataSource_GetMetadataForQueryAndResult(t *testing.T) {
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
	astQueryIDProviderMock := mock_ast.NewMockQueryIDProvider(ctrl)
	astQueryIDProviderMock.EXPECT().GetQueryID(metaQuery.Language, metaQuery.Name, metaQuery.Group).Return(astQueryID, nil)
	similarityIDProviderMock := mock_sast.NewMockSimilarityIDProvider(ctrl)
	similarityIDProviderMock.EXPECT().Calculate(
		gomock.Any(), metaResult.FirstNode.Name, metaResult.FirstNode.Line, metaResult.FirstNode.Column, firstMethodLine,
		gomock.Any(), metaResult.LastNode.Name, metaResult.LastNode.Line, metaResult.LastNode.Column, lastMethodLine,
		astQueryID,
	).Return(similarityID, nil)
	soapAdapterMock := mock_soap.NewMockAdapter(ctrl)
	soapAdapterMock.EXPECT().GetSourcesByScanID(
		scanID,
		[]string{metaResult.FirstNode.FileName, metaResult.LastNode.FileName},
	).Return(&soap.GetSourcesByScanIDResponse{
		GetSourcesByScanIDResult: soap.GetSourcesByScanIDResult{
			CxWSResponseSourcesContent: []soap.CxWSResponseSourcesContent{
				{CxWSResponseSourceContent: soap.CxWSResponseSourceContent{Source: "echo \"hello world 1\""}},
				{CxWSResponseSourceContent: soap.CxWSResponseSourceContent{Source: "echo \"hello world 2\""}},
			},
		},
	}, nil)
	soapAdapterMock.EXPECT().GetResultPathsForQuery(scanID, metaQuery.QueryID).Return(&soap.GetResultPathsForQueryResponse{
		GetResultPathsForQueryResult: soap.GetResultPathsForQueryResult{
			Paths: []soap.ResultPath{
				{
					PathID: metaResult.PathID,
					Nodes: []soap.ResultPathNode{
						{MethodLine: firstMethodLine},
						{MethodLine: "2"},
						{MethodLine: "3"},
						{MethodLine: lastMethodLine},
					},
				},
				{
					PathID: "3",
					Nodes:  []soap.ResultPathNode{{MethodLine: "10"}, {MethodLine: "20"}, {MethodLine: "30"}},
				},
			},
		},
	}, nil)
	metadata := NewMetadataSource(astQueryIDProviderMock, similarityIDProviderMock, soapAdapterMock, tmpDir)

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
