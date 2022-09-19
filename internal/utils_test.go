package internal

import (
	"github.com/checkmarxDev/ast-sast-export/internal/app/export"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDateFromDays(t *testing.T) {
	now := time.Date(2021, 9, 16, 17, 52, 0, 0, time.UTC)
	numDays := 10

	result := GetDateFromDays(numDays, now)

	expected := "2021-9-6"
	assert.Equal(t, expected, result)
}

func TestIsTriageIncluded(t *testing.T) {
	included := []string{export.PresetsOption, export.ResultsOption}
	notIncluded := []string{export.QueriesOption}

	result := IsTriageIncluded(included)
	assert.Equal(t, true, result)

	result = IsTriageIncluded(notIncluded)
	assert.Equal(t, false, result)
}
