package permissions

import (
	"fmt"

	"github.com/checkmarxDev/ast-sast-export/internal/export"
	"github.com/checkmarxDev/ast-sast-export/internal/sliceutils"
	"github.com/dgrijalva/jwt-go"
)

const (
	useOdataPermission           = "use-odata"
	generateScanReportPermission = "generate-scan-report"
	manageAuthProviderPermission = "manage-authentication-providers"
	manageRolesPermission        = "manage-roles"
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
	resultsPermissions := []string{useOdataPermission, generateScanReportPermission}

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

func GetFromJwtClaim(claims jwt.MapClaims, key string) ([]interface{}, error) {
	sastPermissions, exists := claims[key]
	if !exists {
		return make([]interface{}, 0), nil
	}
	multiplePermissions, ok := sastPermissions.([]interface{})
	if ok {
		return multiplePermissions, nil
	}
	singlePermission, ok := sastPermissions.(interface{})
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
