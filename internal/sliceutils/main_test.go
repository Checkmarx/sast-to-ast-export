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
