package internal

import (
	"testing"

	"github.com/dgrijalva/jwt-go"

	"github.com/stretchr/testify/assert"
)

func TestGetPermissionsFromExportOptions(t *testing.T) {
	t.Run("users case", func(t *testing.T) {
		exportOptions := []string{usersExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageRolesPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("teams case", func(t *testing.T) {
		exportOptions := []string{teamsExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results case", func(t *testing.T) {
		exportOptions := []string{resultsExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []interface{}{useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("users+teams case", func(t *testing.T) {
		exportOptions := []string{usersExportOption, teamsExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageRolesPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("users+results case", func(t *testing.T) {
		exportOptions := []string{usersExportOption, resultsExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageRolesPermission, useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results+users case", func(t *testing.T) {
		exportOptions := []string{resultsExportOption, usersExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, manageRolesPermission, useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results+teams case", func(t *testing.T) {
		exportOptions := []string{teamsExportOption, resultsExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []interface{}{manageAuthProviderPermission, useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})
}

func TestGetPermissionsFromJwtClaims(t *testing.T) {
	key := "permissions"

	t.Run("claims without permission", func(t *testing.T) {
		claims := jwt.MapClaims{"test": "test"}

		result, err := getPermissionsFromJwtClaim(claims, key)

		expected := make([]interface{}, 0)
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("claims with one permission", func(t *testing.T) {
		claims := jwt.MapClaims{"test": "test", "permissions": "use-odata"}

		result, err := getPermissionsFromJwtClaim(claims, key)

		expected := []interface{}{"use-odata"}
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("claims with more than one permission", func(t *testing.T) {
		claims := jwt.MapClaims{"test": "test", "permissions": []interface{}{"use-odata", "generate-scan-report"}}

		result, err := getPermissionsFromJwtClaim(claims, key)

		expected := []interface{}{"use-odata", "generate-scan-report"}
		assert.NoError(t, err)
		assert.ElementsMatch(t, expected, result)
	})
}
