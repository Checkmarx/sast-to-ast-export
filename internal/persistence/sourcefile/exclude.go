package sourcefile

import (
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ReadExcludedPaths reads exclude file and returns paths and regex patterns
func ReadExcludedPaths(filePath string) ([]string, []*regexp.Regexp, error) {
	if filePath == "" {
		return nil, nil, nil
	}

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not read exclude file")
	}

	lines := strings.Split(string(fileContent), "\n")
	var excludePaths []string
	var excludePatterns []*regexp.Regexp

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// If line starts and ends with '/', treat it as a regex
		if strings.HasPrefix(line, "/") && strings.HasSuffix(line, "/") {
			pattern := strings.Trim(line, "/")
			regex, err := regexp.Compile(pattern)
			if err != nil {
				log.Warn().Msgf("Invalid regex pattern: %s", pattern)
				continue
			}
			excludePatterns = append(excludePatterns, regex)
		} else {
			// Treat it as a simple string match
			excludePaths = append(excludePaths, line)
		}
	}

	return excludePaths, excludePatterns, nil
}

// IsExcluded checks if a given file is in the exclusion list
func IsExcluded(path string, excludePaths []string) bool {
	for _, exclude := range excludePaths {
		if path == exclude {
			return true
		}
	}
	return false
}
