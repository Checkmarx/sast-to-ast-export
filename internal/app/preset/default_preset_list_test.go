package preset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCases struct {
	id     int
	result bool
}

func TestIsDefaultPreset(t *testing.T) {
	testCases := []TestCases{
		{id: 1, result: true},
		{id: 20, result: true},
		{id: 18, result: false},
		{id: 0, result: false},
		{id: 52, result: true},
		{id: 53, result: false},
	}

	for _, testCase := range testCases {
		result := IsDefaultPreset(testCase.id)
		assert.Equal(t, testCase.result, result)
	}
}
