package resultsmapping

import (
	"fmt"
	"strings"

	"github.com/checkmarxDev/ast-sast-export/internal/app/metadata"
)

type GenerateProvider interface {
	GenerateCSV(record *metadata.Record) [][]string
	WriteAllToSanitizedCsv(records [][]string) []byte
}

func GenerateCSV(record *metadata.Record) [][]string {
	var items [][]string
	items = append(items, []string{
		"result_id",
		"cxone_similarity_id",
		"sast_similarity_id",
	})
	if record == nil {
		return items
	}
	for _, query := range record.Queries {
		for _, result := range query.Results {
			for _, path := range result.Paths {
				items = append(items, []string{
					path.ResultID,
					path.SimilarityID,
					path.SASTSimilarityID,
				})
			}
		}
	}

	return items
}

func WriteAllToSanitizedCsv(records [][]string) []byte {
	for _, row := range records {
		for i := range row {
			row[i] = sanitize(row[i])
		}
	}

	rows := make([]string, len(records))
	for i, record := range records {
		rows[i] = strings.Join(record, ",") + "\n"
	}

	data := strings.Join(rows, "")

	return []byte(data)
}

func sanitize(cell string) string {
	escapedCell := strings.Replace(cell, `"`, `""`, -1)
	return fmt.Sprintf(`"'%s"`, escapedCell)
}
