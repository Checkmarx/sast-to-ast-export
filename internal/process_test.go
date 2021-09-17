package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPermissionsFromExportOptions(t *testing.T) {
	t.Run("users case", func(t *testing.T) {
		exportOptions := []string{usersExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []string{manageAuthProviderPermission, manageRolesPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("teams case", func(t *testing.T) {
		exportOptions := []string{teamsExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []string{manageAuthProviderPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results case", func(t *testing.T) {
		exportOptions := []string{resultsExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []string{useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("users+teams case", func(t *testing.T) {
		exportOptions := []string{usersExportOption, teamsExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []string{manageAuthProviderPermission, manageRolesPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("users+results case", func(t *testing.T) {
		exportOptions := []string{usersExportOption, resultsExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []string{manageAuthProviderPermission, manageRolesPermission, useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results+users case", func(t *testing.T) {
		exportOptions := []string{resultsExportOption, usersExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []string{manageAuthProviderPermission, manageRolesPermission, useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("results+teams case", func(t *testing.T) {
		exportOptions := []string{teamsExportOption, resultsExportOption}
		result := getPermissionsFromExportOptions(exportOptions)

		expected := []string{manageAuthProviderPermission, useOdataPermission, generateScanReportPermission}
		assert.ElementsMatch(t, expected, result)
	})
}
