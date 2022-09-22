package internal

import (
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
