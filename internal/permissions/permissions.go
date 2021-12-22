package permissions

import (
	"fmt"

	"github.com/checkmarxDev/ast-sast-export/internal/export"
	"github.com/checkmarxDev/ast-sast-export/internal/sliceutils"
	"github.com/golang-jwt/jwt"
)

const (
	useOdataPermission           = "use-odata"
	generateScanReportPermission = "generate-scan-report"
	manageAuthProviderPermission = "manage-authentication-providers"
	manageRolesPermission        = "manage-roles"
	viewResults                  = "view-results"
)

var permissionDescription = map[interface{}]string{
	useOdataPermission:           "Sast > API > Use Odata",
	generateScanReportPermission: "Sast > Reports > Generate Scan Report",
	manageAuthProviderPermission: "Access Control > General > Manage Authentication Providers",
	manageRolesPermission:        "Access Control > General > Manage Roles",
}

func GetFromExportOptions(exportOptions []string) []interface{} {
	var output []string

	usersPermissions := []string{manageAuthProviderPermission, manageRolesPermission}
	teamsPermissions := []string{manageAuthProviderPermission}
	resultsPermissions := []string{useOdataPermission, generateScanReportPermission, viewResults}

	for _, exportOption := range exportOptions {
		if exportOption == export.UsersOption {
			output = append(output, usersPermissions...)
		} else if exportOption == export.TeamsOption {
			output = append(output, teamsPermissions...)
		} else if exportOption == export.ResultsOption {
			output = append(output, resultsPermissions...)
		}
	}
	return sliceutils.Unique(sliceutils.ConvertStringToInterface(output))
}

func GetFromJwtClaims(jwtClaims jwt.MapClaims, keys []string) ([]interface{}, error) {
	permissions := make([]interface{}, 0)
	for _, key := range keys {
		claimPermissions, permissionErr := getFromJwtClaim(jwtClaims, key)
		if permissionErr != nil {
			return nil, fmt.Errorf("could not parse %s permissions", key)
		}
		permissions = append(permissions, claimPermissions...)
	}
	return permissions, nil
}

func getFromJwtClaim(claims jwt.MapClaims, key string) ([]interface{}, error) {
	claimValue, exists := claims[key]
	if !exists {
		return make([]interface{}, 0), nil
	}
	multiplePermissions, ok := claimValue.([]interface{})
	if ok {
		return multiplePermissions, nil
	}
	singlePermission, ok := claimValue.(interface{})
	if ok {
		return []interface{}{singlePermission}, nil
	}
	return make([]interface{}, 0), fmt.Errorf("could not parse permissions")
}

func GetDescription(permission interface{}) (string, error) {
	description, ok := permissionDescription[permission]
	if !ok {
		return "", fmt.Errorf("unknown permission %s", permission)
	}
	return description, nil
}

func GetMissing(required, available []interface{}) []interface{} {
	missing := make([]interface{}, 0)
	for _, requiredPermission := range required {
		if !sliceutils.Contains(requiredPermission, available) {
			missing = append(missing, requiredPermission)
		}
	}
	return missing
}
