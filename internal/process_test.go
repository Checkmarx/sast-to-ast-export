package internal

import (
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestGetMissingPermissions(t *testing.T) {
	t.Run("empty lists return empty list", func(t *testing.T) {
		requiredPermissions := make([]interface{}, 0)
		availablePermissions := make([]interface{}, 0)

		result := getMissingPermissions(requiredPermissions, availablePermissions)

		expected := make([]interface{}, 0)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("empty required return empty list", func(t *testing.T) {
		requiredPermissions := make([]interface{}, 0)
		availablePermissions := []interface{}{"a", "b", "c"}

		result := getMissingPermissions(requiredPermissions, availablePermissions)

		expected := make([]interface{}, 0)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("more available than required return empty list", func(t *testing.T) {
		requiredPermissions := []interface{}{"a", "b"}
		availablePermissions := []interface{}{"a", "b", "c"}

		result := getMissingPermissions(requiredPermissions, availablePermissions)

		expected := make([]interface{}, 0)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("missing one returns one item", func(t *testing.T) {
		requiredPermissions := []interface{}{"a", "b", "c"}
		availablePermissions := []interface{}{"a", "b"}

		result := getMissingPermissions(requiredPermissions, availablePermissions)

		expected := []interface{}{"c"}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("missing many returns many items", func(t *testing.T) {
		requiredPermissions := []interface{}{"a", "b", "c", "d", "e", "f"}
		availablePermissions := []interface{}{"a", "b", "d", "f"}

		result := getMissingPermissions(requiredPermissions, availablePermissions)

		expected := []interface{}{"c", "e"}
		assert.ElementsMatch(t, expected, result)
	})
}

func TestGetAvailablePermissions(t *testing.T) {
	claims := jwt.MapClaims{
		"sast-permissions":           []interface{}{"a", "b"},
		"access-control-permissions": []interface{}{"c", "d"},
	}

	result, err := getAvailablePermissions(claims)

	expected := []interface{}{"a", "b", "c", "d"}
	assert.NoError(t, err)
	assert.ElementsMatch(t, expected, result)
}
