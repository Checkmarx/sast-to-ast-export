package common

import (
	"fmt"
	"hash/fnv"
	"strconv"
)

const queryIDIntBase = 10

// GetAstQueryID return queryID for AST
func GetAstQueryID(language, name, group string) (string, error) {
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
