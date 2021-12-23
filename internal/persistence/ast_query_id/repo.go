package ast_query_id

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

//go:embed all_queries.json
var AllQueries string

type Repo struct {
	queries []Query
}

func NewRepo(allQueries string) (*Repo, error) {
	var queries []Query
	unmarshalErr := json.Unmarshal([]byte(allQueries), &queries)
	if unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, "could not unmarshal queries json")
	}
	return &Repo{
		queries,
	}, nil
}

func (e *Repo) GetQueryID(language, name, group string) (string, error) {
	sourcePath := fmt.Sprintf("queries/%s/%s/%s/%s.cs", language, group, name, name)
	for _, query := range e.queries {
		if query.SourcePath == sourcePath {
			return strconv.FormatUint(query.ID, 10), nil
		}
	}
	return "", fmt.Errorf("unknown source path: %s", sourcePath)
}
