package metadata

import (
	"sort"
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"

	mock_app_ast_query_id "github.com/checkmarxDev/ast-sast-export/test/mocks/app/ast_query_id"
	mock_app_method_line "github.com/checkmarxDev/ast-sast-export/test/mocks/app/method_line"
	mock_app_source_file "github.com/checkmarxDev/ast-sast-export/test/mocks/app/source_file"

	mock_integration_similarity "github.com/checkmarxDev/ast-sast-export/test/mocks/integration/similarity"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type testResultData struct {
	MethodLines  []string
	SimilarityID string
}

func TestMetadataFactory_GetMetadataForQueryAndResult(t *testing.T) {
	scanID := "1000001"
	astQueryID := "12532796926860742976"
	similarityID1 := "-1234567890"
	similarityID2 := "-1234567891"
	metaResult1Data := testResultData{
		SimilarityID: "1234567890",
		MethodLines:  []string{"100", "2", "3", "101"},
	}
	metaResult1 := Result{
		PathID:       "2",
		ResultID:     "1000002",
		SimilarityID: similarityID1,
		FirstNode: Node{
			FileName: "Goatlin-develop/packages/clients/android/app/src/main/java/com/cx/goatlin/EditNoteActivity.kt",
			Name:     "text",
			Line:     "83",
			Column:   "78",
		},
		LastNode: Node{
			FileName: "Goatlin-develop/packages/clients/android/app/src/main/java/com/cx/goatlin/helpers/DatabaseHelper.kt",
			Name:     "note",
			Line:     "129",
			Column:   "28",
		},
	}
	metaResult2Data := testResultData{
		SimilarityID: "9492845843",
		MethodLines:  []string{"43", "21", "13"},
	}
	metaResult2 := Result{
		PathID:       "3",
		ResultID:     "1000002",
		SimilarityID: similarityID2,
		FirstNode: Node{
			FileName: "path/file1.kt",
			Name:     "text",
			Line:     "83",
			Column:   "78",
		},
		LastNode: Node{
			FileName: "path/file2.kt",
			Name:     "note",
			Line:     "129",
			Column:   "28",
		},
	}
	metaQuery := &Query{
		QueryID:  "6300",
		Language: "Kotlin",
		Name:     "SQL_Injection",
		Group:    "Kotlin_High_Risk",
		Results:  []*Result{&metaResult1, &metaResult2},
	}

	ctrl := gomock.NewController(t)
	tmpDir := t.TempDir()
	astQueryIDProviderMock := mock_app_ast_query_id.NewMockASTQueryIDProvider(ctrl)
	astQueryIDProviderMock.EXPECT().GetQueryID(metaQuery.Language, metaQuery.Name, metaQuery.Group, metaQuery.QueryID).Return(astQueryID, nil)
	similarityIDProviderMock := mock_integration_similarity.NewMockIDProvider(ctrl)
	similarityIDProviderMock.EXPECT().Calculate(
		gomock.Any(), metaResult1.FirstNode.Name, metaResult1.FirstNode.Line, metaResult1.FirstNode.Column, metaResult1Data.MethodLines[0],
		gomock.Any(), metaResult1.LastNode.Name, metaResult1.LastNode.Line, metaResult1.LastNode.Column, metaResult1Data.MethodLines[3],
		astQueryID, 0,
	).Return(metaResult1Data.SimilarityID, nil)
	similarityIDProviderMock.EXPECT().Calculate(
		gomock.Any(), metaResult2.FirstNode.Name, metaResult2.FirstNode.Line, metaResult2.FirstNode.Column, metaResult2Data.MethodLines[0],
		gomock.Any(), metaResult2.LastNode.Name, metaResult2.LastNode.Line, metaResult2.LastNode.Column, metaResult2Data.MethodLines[2],
		astQueryID, 0,
	).Return(metaResult2Data.SimilarityID, nil)
	sourceProviderMock := mock_app_source_file.NewMockSourceFileRepo(ctrl)
	sourceProviderMock.EXPECT().
		DownloadSourceFiles(scanID, gomock.Any(), gomock.Eq("")).
		DoAndReturn(
			func(_ string, files []interfaces.SourceFile, rmvdir string) error {
				expectedFiles := []string{
					metaResult1.FirstNode.FileName,
					metaResult1.LastNode.FileName,
					metaResult2.FirstNode.FileName,
					metaResult2.LastNode.FileName,
				}
				var result []string
				for _, v := range files {
					result = append(result, v.RemoteName)
				}
				assert.ElementsMatch(t, expectedFiles, result)
				assert.Equal(t, "", rmvdir)
				return nil
			},
		)
	methodLineProvider := mock_app_method_line.NewMockMethodLineRepo(ctrl)
	methodLinesResult := []*interfaces.ResultPath{
		{PathID: metaResult1.PathID, MethodLines: metaResult1Data.MethodLines},
		{PathID: metaResult2.PathID, MethodLines: metaResult2Data.MethodLines},
	}
	methodLineProvider.EXPECT().
		GetMethodLinesByPath(scanID, metaQuery.QueryID).
		Return(methodLinesResult, nil)
	metadata := NewMetadataFactory(astQueryIDProviderMock, similarityIDProviderMock, sourceProviderMock, methodLineProvider, tmpDir, 0, "")

	result, err := metadata.GetMetadataRecord(scanID, []*Query{metaQuery})
	assert.NoError(t, err)

	expectedResult := &Record{
		Queries: []*RecordQuery{
			{
				QueryID: metaQuery.QueryID,
				Results: []*RecordResult{
					{
						ResultID: metaResult1.ResultID,
						Paths: []*RecordPath{
							{
								PathID:           metaResult1.PathID,
								SimilarityID:     metaResult1Data.SimilarityID,
								ResultID:         metaResult1.ResultID,
								SASTSimilarityID: similarityID1,
							},
							{
								PathID:           metaResult2.PathID,
								SimilarityID:     metaResult2Data.SimilarityID,
								ResultID:         metaResult2.ResultID,
								SASTSimilarityID: similarityID2,
							},
						},
					},
				},
			},
		},
	}

	sortRecordPaths(expectedResult)
	sortRecordPaths(result)

	assert.Equal(t, expectedResult, result)
}

func sortRecordPaths(record *Record) {
	for _, query := range record.Queries {
		for _, result := range query.Results {
			sort.Slice(result.Paths, func(i, j int) bool {
				return result.Paths[i].PathID < result.Paths[j].PathID
			})
		}
	}
}
