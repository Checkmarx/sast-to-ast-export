package permissions

import (
	"fmt"
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/app/export"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestGetFromExportOptions(t *testing.T) {
	t.Run("users case", func(t *testing.T) {
		exportOptions := []string{export.UsersOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageRolesPermission, manageSystemSettings}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("teams case", func(t *testing.T) {
		exportOptions := []string{export.TeamsOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageSystemSettings}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results case", func(t *testing.T) {
		exportOptions := []string{export.ResultsOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{useOdataPermission, generateScanReportPermission, viewResults, manageSystemSettings}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("users+teams case", func(t *testing.T) {
		exportOptions := []string{export.UsersOption, export.TeamsOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageRolesPermission, manageSystemSettings}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("users+results case", func(t *testing.T) {
		exportOptions := []string{export.UsersOption, export.ResultsOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{
			manageAuthProviderPermission, manageRolesPermission, useOdataPermission, generateScanReportPermission,
			viewResults, manageSystemSettings,
		}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results+users case", func(t *testing.T) {
		exportOptions := []string{export.ResultsOption, export.UsersOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{
			manageAuthProviderPermission, manageRolesPermission, useOdataPermission, generateScanReportPermission,
			viewResults, manageSystemSettings,
		}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results+teams case", func(t *testing.T) {
		exportOptions := []string{export.TeamsOption, export.ResultsOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{
			manageAuthProviderPermission, useOdataPermission, generateScanReportPermission, viewResults,
			manageSystemSettings,
		}
		assert.ElementsMatch(t, expected, result)
	})
}

func TestGetAllFromJwtClaims(t *testing.T) {
	claims := jwt.MapClaims{
		"aaa": []interface{}{"a", "b"},
		"bbb": []interface{}{"c", "d"},
		"ccc": []interface{}{"e", "f"},
	}

	result, err := GetFromJwtClaims(claims, []string{"aaa", "bbb"})

	expected := []interface{}{"a", "b", "c", "d"}
	assert.NoError(t, err)
	assert.ElementsMatch(t, expected, result)
}

func TestGetFromJwtClaims(t *testing.T) {
	key := "permissions"

	t.Run("claims without permission", func(t *testing.T) {
		claims := jwt.MapClaims{"test": "test"}

		result, err := getFromJwtClaim(claims, key)

		expected := make([]interface{}, 0)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("claims with one permission", func(t *testing.T) {
		claims := jwt.MapClaims{"test": "test", "permissions": "use-odata"}

		result, err := getFromJwtClaim(claims, key)

		expected := []interface{}{"use-odata"}
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("claims with more than one permission", func(t *testing.T) {
		claims := jwt.MapClaims{"test": "test", "permissions": []interface{}{"use-odata", "generate-scan-report"}}

		result, err := getFromJwtClaim(claims, key)

		expected := []interface{}{"use-odata", "generate-scan-report"}
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})
}

func TestGetMissing(t *testing.T) {
	t.Run("empty lists return empty list", func(t *testing.T) {
		required := make([]interface{}, 0)
		available := make([]interface{}, 0)

		result := GetMissing(required, available)

		expected := make([]interface{}, 0)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("empty required return empty list", func(t *testing.T) {
		required := make([]interface{}, 0)
		available := []interface{}{"a", "b", "c"}

		result := GetMissing(required, available)

		expected := make([]interface{}, 0)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("more available than required return empty list", func(t *testing.T) {
		required := []interface{}{"a", "b"}
		available := []interface{}{"a", "b", "c"}

		result := GetMissing(required, available)

		expected := make([]interface{}, 0)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("missing one returns one item", func(t *testing.T) {
		required := []interface{}{"a", "b", "c"}
		available := []interface{}{"a", "b"}

		result := GetMissing(required, available)

		expected := []interface{}{"c"}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("missing many returns many items", func(t *testing.T) {
		required := []interface{}{"a", "b", "c", "d", "e", "f"}
		available := []interface{}{"a", "b", "d", "f"}

		result := GetMissing(required, available)

		expected := []interface{}{"c", "e"}
		assert.ElementsMatch(t, expected, result)
	})
}

func TestGetDescription(t *testing.T) {
	successTests := [][]string{
		{useOdataPermission, "Sast > API > Use Odata"},
		{generateScanReportPermission, "Sast > Reports > Generate Scan Report"},
		{viewResults, "Sast > Scan Results > View Results"},
		{manageAuthProviderPermission, "Access Control > General > Manage Authentication Providers"},
		{manageRolesPermission, "Access Control > General > Manage Roles"},
	}
	for _, test := range successTests {
		testName := fmt.Sprintf("succeeds if %s exists", test[0])
		t.Run(testName, func(t *testing.T) {
			result, err := GetDescription(test[0])

			assert.NoError(t, err)
			assert.Equal(t, test[1], result)
		})
	}

	t.Run("fails if permission doesn't exist", func(t *testing.T) {
		permissionID := "invalid"

		result, err := GetDescription(permissionID)

		assert.Error(t, err)
		assert.Equal(t, err.Error(), "unknown permission invalid")
		assert.Equal(t, "", result)
	})
}
