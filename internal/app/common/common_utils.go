package common

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const queryIDIntBase = 10

type RenameEntry struct {
	OldPath      string `json:"oldPath"`
	NewPath      string `json:"newPath"`
	OldPathAstID uint64 `json:"oldPathAstID"`
	NewPathAstID uint64 `json:"newPathAstID"`
}

var renameMap map[string]uint64

// GetAstQueryID returns the query ID for AST
func GetAstQueryID(language, name, group string) (string, error) {
	if renameMap == nil {
		return "", fmt.Errorf("rename data not loaded or initialization failed")
	}

	sourcePath := fmt.Sprintf("queries/%s/%s/%s/%s.cs", language, group, name, name)

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

// LoadRename attempts to load the rename mapping from the specified source (file path or URL).
// This function should be called ONCE during application initialization
func LoadRename(renameSource string) error {
	var data []byte
	var err error

	if renameSource == "" {
		return fmt.Errorf("rename source (file path or URL) cannot be empty")
	}

	// Check if it's a URL or a local file path
	if strings.HasPrefix(renameSource, "http://") || strings.HasPrefix(renameSource, "https://") {
		// Load from URL
		resp, httpErr := http.Get(renameSource) //nolint:gosec
		if httpErr != nil {
			return fmt.Errorf("failed to fetch rename data from URL '%s': %w", renameSource, httpErr)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to fetch rename data from URL '%s': HTTP %d, Body: %s", renameSource, resp.StatusCode, string(bodyBytes))
		}

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body from URL '%s': %w", renameSource, err)
		}
	} else {
		// Load from local file
		data, err = os.ReadFile(renameSource)
		if err != nil {
			return fmt.Errorf("failed to read rename file '%s': %w", renameSource, err)
		}
	}

	var entries []RenameEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to parse rename JSON data from source '%s': %w", renameSource, err)
	}

	currentRenameMap := make(map[string]uint64)
	for _, entry := range entries {
		currentRenameMap[entry.NewPath] = entry.OldPathAstID
	}

	renameMap = currentRenameMap

	return nil
}
