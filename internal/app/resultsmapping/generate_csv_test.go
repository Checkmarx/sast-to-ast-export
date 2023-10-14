package resultsmapping

import (
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/app/metadata"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCsv(t *testing.T) {
	var latestVersionHeaders = []string{"result_id", "cxone_similarity_id", "sast_similarity_id"}
	similarityID1New := "-1234567890"
	similarityID2New := "-1234567891"
	similarityID1 := "-1234567890"
	similarityID2 := "-1234567891"
	metaResult1 := metadata.Result{
		PathID:       "2",
		ResultID:     "1000002",
		SimilarityID: similarityID1,
		FirstNode: metadata.Node{
			FileName: "Goatlin-develop/packages/clients/android/app/src/main/java/com/cx/goatlin/EditNoteActivity.kt",
			Name:     "text",
			Line:     "83",
			Column:   "78",
		},
		LastNode: metadata.Node{
			FileName: "Goatlin-develop/packages/clients/android/app/src/main/java/com/cx/goatlin/helpers/DatabaseHelper.kt",
			Name:     "note",
			Line:     "129",
			Column:   "28",
		},
	}
	metaResult2 := metadata.Result{
		PathID:       "3",
		ResultID:     "1000002",
		SimilarityID: similarityID2,
		FirstNode: metadata.Node{
			FileName: "path/file1.kt",
			Name:     "text",
			Line:     "83",
			Column:   "78",
		},
		LastNode: metadata.Node{
			FileName: "path/file2.kt",
			Name:     "note",
			Line:     "129",
			Column:   "28",
		},
	}
	metaQuery := &metadata.Query{
		QueryID:  "6300",
		Language: "Kotlin",
		Name:     "SQL_Injection",
		Group:    "Kotlin_High_Risk",
		Results:  []*metadata.Result{&metaResult1, &metaResult2},
	}
	inputRecord1 := &metadata.Record{
		Queries: []*metadata.RecordQuery{
			{
				QueryID: metaQuery.QueryID,
				Results: []*metadata.RecordResult{
					{
						ResultID: metaResult1.ResultID,
						Paths: []*metadata.RecordPath{
							{
								PathID:           metaResult1.PathID,
								SimilarityID:     similarityID1New,
								SASTSimilarityID: similarityID1,
							},
							{
								PathID:           metaResult2.PathID,
								SimilarityID:     similarityID2New,
								SASTSimilarityID: similarityID2,
							},
						},
					},
				},
			},
		},
	}
	inputRecord2 := &metadata.Record{
		Queries: []*metadata.RecordQuery{
			{
				QueryID: metaQuery.QueryID,
				Results: []*metadata.RecordResult{
					{
						ResultID: metaResult1.ResultID,
						Paths: []*metadata.RecordPath{
							{
								PathID:           metaResult1.PathID,
								SimilarityID:     similarityID1New,
								SASTSimilarityID: similarityID1,
							},
						},
					},
				},
			},
		},
	}

	allRecords := []*metadata.Record{
		inputRecord1,
		inputRecord2,
	}

	items1 := [][]string{
		{"result_id", "cxone_similarity_id", "sast_similarity_id"},
		{"", "-1234567890", "-1234567890"},
		{"", "-1234567891", "-1234567891"},
	}

	items2 := [][]string{
		{"result_id", "cxone_similarity_id", "sast_similarity_id"},
		{"", "-1234567890", "-1234567890"},
	}

	itemsAll := [][]string{
		{"result_id", "cxone_similarity_id", "sast_similarity_id"},
		{"", "-1234567890", "-1234567890"},
		{"", "-1234567891", "-1234567891"},
		{"", "-1234567890", "-1234567890"},
	}

	t.Run("validate csv data returned from results", func(t *testing.T) {
		result := GenerateCSV([]*metadata.Record{inputRecord1})

		assert.Equal(t, items1, result)
	})

	t.Run("validate csv data returned from results", func(t *testing.T) {
		result := GenerateCSV([]*metadata.Record{inputRecord2})

		assert.Equal(t, items2, result)
	})

	t.Run("validate csv data returned from results", func(t *testing.T) {
		result := GenerateCSV(allRecords)

		assert.Equal(t, itemsAll, result)
	})

	t.Run("success returns only headers if not exists data to generate csv from results", func(t *testing.T) {
		expectedResult := [][]string{latestVersionHeaders}
		result := GenerateCSV([]*metadata.Record{})

		assert.Equal(t, expectedResult, result)
	})

	t.Run("success returns only headers if is passed nil from model", func(t *testing.T) {
		expectedResult := [][]string{latestVersionHeaders}
		result := GenerateCSV(nil)

		assert.Equal(t, expectedResult, result)
	})
}

func TestWriteAllToSanitizedCsv(t *testing.T) {
	items := [][]string{
		{"result_id", "path_id", "cxone_similarity_id", "sast_similarity_id"},
		{"", "2", "-1234567890", "-1234567890"},
		{"", "3", "-1234567891", "-1234567891"},
	}

	t.Run("should return a byte array of sanitize records", func(t *testing.T) {
		expected := []byte(`"'result_id","'path_id","'cxone_similarity_id","'sast_similarity_id"` + "\n" + `"'","'2",` +
			`"'-1234567890","'-1234567890"` + "\n" + `"'","'3","'-1234567891","'-1234567891"` + "\n")

		result := WriteAllToSanitizedCsv(items)
		assert.Equal(t, expected, result)
	})

	t.Run("should return an empty byte array if there are no records", func(t *testing.T) {
		expected := []byte("")

		result := WriteAllToSanitizedCsv([][]string{})
		assert.Equal(t, expected, result)
	})
}

func TestSanitize(t *testing.T) {
	t.Run("with single quote in cell", func(t *testing.T) {
		input := `=1+2'" ;,=1+2`
		expectedResult := `"'=1+2'"" ;,=1+2"`
		result := sanitize(input)

		assert.Equal(t, expectedResult, result)
	})

	t.Run("without single quote in cell", func(t *testing.T) {
		input := `=1+2";=1+2`
		expectedResult := `"'=1+2"";=1+2"`
		result := sanitize(input)

		assert.Equal(t, expectedResult, result)
	})
}
