package sliceutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	heap := []interface{}{"bar", "baz"}

	assert.False(t, Contains("foo", heap))
	assert.True(t, Contains("bar", heap))
}

func TestUnique(t *testing.T) {
	heap := []interface{}{"a", "b", "c", "c", "d", "e", "f", "f", "g"}

	result := Unique(heap)

	expected := []interface{}{"a", "b", "c", "d", "e", "f", "g"}
	assert.Equal(t, expected, result)
}

func TestConvertStringToInterface(t *testing.T) {
	heap := []string{"a", "b", "c"}

	result := ConvertStringToInterface(heap)

	expected := []interface{}{"a", "b", "c"}
	assert.Equal(t, expected, result)
}

func TestConvertInterfaceToString(t *testing.T) {
	heap := []interface{}{"a", "b", "c"}

	result := ConvertInterfaceToString(heap)

	expected := []string{"a", "b", "c"}
	assert.Equal(t, expected, result)
}
