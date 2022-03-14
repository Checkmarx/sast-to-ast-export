package astquery

import (
	"fmt"
	"hash/fnv"
	"strconv"
)

const queryIDIntBase = 10

type Provider struct{}

func NewProvider() (*Provider, error) {
	return &Provider{}, nil
}

func (e *Provider) GetQueryID(language, name, group string) (string, error) {
	sourcePath := fmt.Sprintf("queries/%s/%s/%s/%s.cs", language, group, name, name)
	queryID, queryIDErr := hash(sourcePath)
	if queryIDErr != nil {
		return "", queryIDErr
	}
	return strconv.FormatUint(queryID, queryIDIntBase), nil
}

func hash(s string) (uint64, error) {
	h := fnv.New64()
	_, err := h.Write([]byte(s))
	return h.Sum64(), err
}
