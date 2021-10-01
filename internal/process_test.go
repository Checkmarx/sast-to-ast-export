package internal

import (
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/export"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

type validatePermissionTest struct {
	JwtClaims     jwt.MapClaims
	ExportOptions []string
	ExpectErr     bool
	Message       string
}

func TestValidatePermissions(t *testing.T) {
	tests := []validatePermissionTest{
		{jwt.MapClaims{}, []string{}, false, "empty claims and export options"},
		{jwt.MapClaims{}, []string{export.UsersOption}, true, "empty claims"},
		{
			jwt.MapClaims{"access-control-permissions": "manage-authentication-providers"},
			[]string{export.TeamsOption},
			false,
			"single, correct permission",
		},
		{
			jwt.MapClaims{"access-control-permissions": "invalid"},
			[]string{export.TeamsOption},
			true,
			"single, incorrect permission",
		},
		{
			jwt.MapClaims{"access-control-permissions": nil},
			[]string{export.TeamsOption},
			true,
			"single, invalid permission",
		},
		{
			jwt.MapClaims{"access-control-permissions": "manage-authentication-providers"},
			[]string{export.UsersOption},
			true,
			"missing one permission",
		},
		{
			jwt.MapClaims{"access-control-permissions": []interface{}{"manage-authentication-providers", "manage-roles"}},
			[]string{export.UsersOption},
			false,
			"permission list with correct permissions",
		},
		{
			jwt.MapClaims{"access-control-permissions": []interface{}{"invalid", "manage-roles"}},
			[]string{export.UsersOption},
			true,
			"permission list with incorrect permissions",
		},
		{
			jwt.MapClaims{
				"sast-permissions":           []interface{}{"use-odata", "generate-scan-report"},
				"access-control-permissions": []interface{}{"manage-roles", "manage-authentication-providers"},
			},
			[]string{export.UsersOption, export.ResultsOption},
			false,
			"multiple permission lists with correct permissions",
		},
		{
			jwt.MapClaims{
				"sast-permissions":           []interface{}{"invalid", "generate-scan-report"},
				"access-control-permissions": []interface{}{"manage-roles", "manage-authentication-providers"},
			},
			[]string{export.UsersOption, export.ResultsOption},
			true,
			"multiple permission lists with incorrect permissions",
		},
	}
	for _, test := range tests {
		err := validatePermissions(test.JwtClaims, test.ExportOptions)

		if test.ExpectErr {
			assert.Error(t, err, test.Message)
		} else {
			assert.NoError(t, err, test.Message)
		}
	}
}
