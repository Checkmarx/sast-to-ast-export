package permissions

import (
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/export"

	"github.com/dgrijalva/jwt-go"

	"github.com/stretchr/testify/assert"
)

func TestGetFromExportOptions(t *testing.T) {
	t.Run("users case", func(t *testing.T) {
		exportOptions := []string{export.UsersOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageRolesPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("teams case", func(t *testing.T) {
		exportOptions := []string{export.TeamsOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results case", func(t *testing.T) {
		exportOptions := []string{export.ResultsOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("users+teams case", func(t *testing.T) {
		exportOptions := []string{export.UsersOption, export.TeamsOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageRolesPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("users+results case", func(t *testing.T) {
		exportOptions := []string{export.UsersOption, export.ResultsOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageRolesPermission, useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results+users case", func(t *testing.T) {
		exportOptions := []string{export.ResultsOption, export.UsersOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageRolesPermission, useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results+teams case", func(t *testing.T) {
		exportOptions := []string{export.TeamsOption, export.ResultsOption}
		result := GetFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})
}

func TestGetFromJwtClaims(t *testing.T) {
	key := "permissions"

	t.Run("claims without permission", func(t *testing.T) {
		claims := jwt.MapClaims{"test": "test"}

		result, err := GetFromJwtClaim(claims, key)

		expected := make([]interface{}, 0)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("claims with one permission", func(t *testing.T) {
		claims := jwt.MapClaims{"test": "test", "permissions": "use-odata"}

		result, err := GetFromJwtClaim(claims, key)

		expected := []interface{}{"use-odata"}
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("claims with more than one permission", func(t *testing.T) {
		claims := jwt.MapClaims{"test": "test", "permissions": []interface{}{"use-odata", "generate-scan-report"}}

		result, err := GetFromJwtClaim(claims, key)

		expected := []interface{}{"use-odata", "generate-scan-report"}
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})
}
