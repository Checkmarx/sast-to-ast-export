package common

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"strconv"
	"sync"
)

const queryIDIntBase = 10

type RenameEntry struct {
	OldPath      string `json:"oldPath"`
	NewPath      string `json:"newPath"`
	OldPathAstID uint64 `json:"oldPathAstID"`
	NewPathAstID uint64 `json:"newPathAstID"`
}

var (
	renameMap map[string]uint64
	loadOnce  sync.Once
	loadErr   error
)

const engineConfigURL = "https://raw.githubusercontent.com/Checkmarx/sast-to-ast-export/refs/heads/master/data/renames.json"

// GetAstQueryID returns the query ID for AST
func GetAstQueryID(language, name, group string) (string, error) {
	// Ensure the rename data is loaded before accessing it
	if err := LoadRename(); err != nil {
		return "", err
	}

	sourcePath := fmt.Sprintf("queries/%s/%s/%s/%s.cs", language, group, name, name)

	// Check if path is in renames.json
	if oldAstID, exists := renameMap[sourcePath]; exists {
		return strconv.FormatUint(oldAstID, queryIDIntBase), nil
	}

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

// Overwritten queries that have been renamed between sast versions need to have the original AST ID
// And AST ID is calculated by hashing the path of the query, this means we need to store the old AST ID for the new path
func LoadRename() error {
	// Load renames.json from engineConfigURL, ensuring it runs only once
	loadOnce.Do(func() {
		resp, err := http.Get(engineConfigURL)
		if err != nil {
			loadErr = fmt.Errorf("failed to fetch rename data: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			loadErr = fmt.Errorf("failed to fetch rename data: HTTP %d", resp.StatusCode)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			loadErr = fmt.Errorf("failed to read response body: %w", err)
			return
		}

		var entries []RenameEntry
		if err := json.Unmarshal(body, &entries); err != nil {
			loadErr = fmt.Errorf("failed to parse JSON: %w", err)
			return
		}

		renameMap = make(map[string]uint64)
		for _, entry := range entries {
			renameMap[entry.NewPath] = entry.OldPathAstID
		}
	})

	return loadErr
}
