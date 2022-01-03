package ast_query_id

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/pkg/errors"
)

//go:embed all_queries.json
var AllQueries string

const (
	sourcePathRegex = "queries/([^/]+)/([^/]+)/([^/]+)/.+\\.cs"
)

type Repo struct {
	queries map[string][]Query
}

func NewRepo(allQueries string) (*Repo, error) {
	var queries []Query
	unmarshalErr := json.Unmarshal([]byte(allQueries), &queries)
	if unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, "could not unmarshal queries json")
	}
	r, regexErr := regexp.Compile(sourcePathRegex)
	if regexErr != nil {
		return nil, errors.Wrap(regexErr, "could not compile source path regex")
	}
	i := Repo{queries: map[string][]Query{}}
	for _, q := range queries {
		match := r.FindStringSubmatch(q.SourcePath)
		language := match[1]
		if _, ok := i.queries[language]; !ok {
			i.queries[language] = []Query{}
		}
		i.queries[language] = append(i.queries[language], q)
	}
	return &i, nil
}

func (e *Repo) GetQueryID(language, name, group string) (string, error) {
	sourcePath := fmt.Sprintf("queries/%s/%s/%s/%s.cs", language, group, name, name)
	for _, query := range e.queries[language] {
		if query.SourcePath == sourcePath {
			return strconv.FormatUint(query.ID, 10), nil
		}
	}
	return "", fmt.Errorf("unknown source path: %s", sourcePath)
}

func (e *Repo) GetAllQueryIDsByGroup(language, name string) ([]interfaces.ASTQuery, error) {
	pattern := fmt.Sprintf("queries/%s/([^/]+)/%s/%s.cs", language, name, name)
	r, regexErr := regexp.Compile(pattern)
	if regexErr != nil {
		return nil, regexErr
	}
	var out []interfaces.ASTQuery
	for _, query := range e.queries[language] {
		if r.MatchString(query.SourcePath) {
			match := r.FindStringSubmatch(query.SourcePath)
			queryID := strconv.FormatUint(query.ID, 10)
			out = append(out, interfaces.ASTQuery{
				Language: language,
				Group:    match[1],
				Name:     name,
				QueryID:  queryID,
			})
		}
	}
	return out, nil
}
