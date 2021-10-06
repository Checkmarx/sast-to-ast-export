package internal

import (
	"fmt"
	export2 "github.com/checkmarxDev/ast-sast-export/internal/test/mocks/export"
	sast2 "github.com/checkmarxDev/ast-sast-export/internal/test/mocks/sast"
	"github.com/golang/mock/gomock"
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

type fetchUsersExpectProp struct {
	ReturnError error
	RunCount    int
}

type fetchUsersExpect struct {
	GetUsers                 fetchUsersExpectProp
	GetRoles                 fetchUsersExpectProp
	GetLdapRoleMappings      fetchUsersExpectProp
	GetSamlRoleMappings      fetchUsersExpectProp
	GetLdapServers           fetchUsersExpectProp
	GetSamlIdentityProviders fetchUsersExpectProp
}

type writeUsersExpectProp struct {
	ReturnError error
	RunCount    int
}

type writeUsersExpect struct {
	Users                 writeUsersExpectProp
	Roles                 writeUsersExpectProp
	LdapRoleMappings      writeUsersExpectProp
	SamlRoleMappings      writeUsersExpectProp
	LdapServers           writeUsersExpectProp
	SamlIdentityProviders writeUsersExpectProp
}

func fetchUsersSetupExpects(client *sast2.MockClient, expect *fetchUsersExpect) {
	client.EXPECT().
		GetUsers().
		Return([]byte{}, expect.GetUsers.ReturnError).
		MinTimes(expect.GetUsers.RunCount).
		MaxTimes(expect.GetUsers.RunCount)
	client.EXPECT().
		GetRoles().
		Return([]byte{}, expect.GetRoles.ReturnError).
		MinTimes(expect.GetRoles.RunCount).
		MaxTimes(expect.GetRoles.RunCount)
	client.EXPECT().
		GetLdapRoleMappings().
		Return([]byte{}, expect.GetLdapRoleMappings.ReturnError).
		MinTimes(expect.GetLdapRoleMappings.RunCount).
		MaxTimes(expect.GetLdapRoleMappings.RunCount)
	client.EXPECT().
		GetSamlRoleMappings().
		Return([]byte{}, expect.GetSamlRoleMappings.ReturnError).
		MinTimes(expect.GetSamlRoleMappings.RunCount).
		MaxTimes(expect.GetSamlRoleMappings.RunCount)
	client.EXPECT().
		GetLdapServers().
		Return([]byte{}, expect.GetLdapServers.ReturnError).
		MinTimes(expect.GetLdapServers.RunCount).
		MaxTimes(expect.GetLdapServers.RunCount)
	client.EXPECT().
		GetSamlIdentityProviders().
		Return([]byte{}, expect.GetSamlIdentityProviders.ReturnError).
		MinTimes(expect.GetSamlIdentityProviders.RunCount).
		MaxTimes(expect.GetSamlIdentityProviders.RunCount)
}

func writeUsersSetupExpects(exporter *export2.MockExporter, expect *writeUsersExpect) {
	exporter.EXPECT().
		AddFileWithDataSource(export.UsersFileName, gomock.Any()).
		DoAndReturn(func(_ string, _ func() ([]byte, error)) error {
			return expect.Users.ReturnError
		}).
		MinTimes(expect.Users.RunCount).
		MaxTimes(expect.Users.RunCount)
	exporter.EXPECT().
		AddFileWithDataSource(export.RolesFileName, gomock.Any()).
		DoAndReturn(func(_ string, _ func() ([]byte, error)) error {
			return expect.Roles.ReturnError
		}).
		MinTimes(expect.Roles.RunCount).
		MaxTimes(expect.Roles.RunCount)
	exporter.EXPECT().
		AddFileWithDataSource(export.LdapRoleMappingsFileName, gomock.Any()).
		DoAndReturn(func(_ string, _ func() ([]byte, error)) error {
			return expect.LdapRoleMappings.ReturnError
		}).
		MinTimes(expect.LdapRoleMappings.RunCount).
		MaxTimes(expect.LdapRoleMappings.RunCount)
	exporter.EXPECT().
		AddFileWithDataSource(export.SamlRoleMappingsFileName, gomock.Any()).
		DoAndReturn(func(_ string, _ func() ([]byte, error)) error {
			return expect.SamlRoleMappings.ReturnError
		}).
		MinTimes(expect.SamlRoleMappings.RunCount).
		MaxTimes(expect.SamlRoleMappings.RunCount)
	exporter.EXPECT().
		AddFileWithDataSource(export.LdapServersFileName, gomock.Any()).
		DoAndReturn(func(_ string, _ func() ([]byte, error)) error {
			return expect.LdapServers.ReturnError
		}).
		MinTimes(expect.LdapServers.RunCount).
		MaxTimes(expect.LdapServers.RunCount)
	exporter.EXPECT().
		AddFileWithDataSource(export.SamlIdpFileName, gomock.Any()).
		DoAndReturn(func(_ string, _ func() ([]byte, error)) error {
			return expect.SamlIdentityProviders.ReturnError
		}).
		MinTimes(expect.SamlIdentityProviders.RunCount).
		MaxTimes(expect.SamlIdentityProviders.RunCount)
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

//nolint:funlen
func TestFetchUsersData(t *testing.T) {
	t.Run("fails if fetch users fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to read users")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers: fetchUsersExpectProp{err, 1},
		})
		exporter.EXPECT().
			AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()

		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if add users file fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to write users file")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers: fetchUsersExpectProp{nil, 1},
		})
		writeUsersSetupExpects(exporter, &writeUsersExpect{
			Users: writeUsersExpectProp{err, 1},
		})
		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if fetch roles fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to read roles")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers: fetchUsersExpectProp{nil, 1},
			GetRoles: fetchUsersExpectProp{err, 1},
		})
		exporter.EXPECT().
			AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()

		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if add roles file fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to write roles file")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers: fetchUsersExpectProp{nil, 1},
			GetRoles: fetchUsersExpectProp{nil, 1},
		})
		writeUsersSetupExpects(exporter, &writeUsersExpect{
			Users: writeUsersExpectProp{nil, 1},
			Roles: writeUsersExpectProp{err, 1},
		})
		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if fetch LDAP role mappings fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to read LDAP role mappings")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers:            fetchUsersExpectProp{nil, 1},
			GetRoles:            fetchUsersExpectProp{nil, 1},
			GetLdapRoleMappings: fetchUsersExpectProp{err, 1},
		})
		exporter.EXPECT().
			AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()

		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if add LDAP role mappings file fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to write LDAP role mappings file")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers:            fetchUsersExpectProp{nil, 1},
			GetRoles:            fetchUsersExpectProp{nil, 1},
			GetLdapRoleMappings: fetchUsersExpectProp{nil, 1},
		})
		writeUsersSetupExpects(exporter, &writeUsersExpect{
			Users:            writeUsersExpectProp{nil, 1},
			Roles:            writeUsersExpectProp{nil, 1},
			LdapRoleMappings: writeUsersExpectProp{err, 1},
		})
		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if fetch SAML role mappings fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to read SAML role mappings")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers:            fetchUsersExpectProp{nil, 1},
			GetRoles:            fetchUsersExpectProp{nil, 1},
			GetLdapRoleMappings: fetchUsersExpectProp{nil, 1},
			GetSamlRoleMappings: fetchUsersExpectProp{err, 1},
		})
		exporter.EXPECT().
			AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()

		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if add SAML role mappings file fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to write SAML role mappings file")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers:            fetchUsersExpectProp{nil, 1},
			GetRoles:            fetchUsersExpectProp{nil, 1},
			GetLdapRoleMappings: fetchUsersExpectProp{nil, 1},
			GetSamlRoleMappings: fetchUsersExpectProp{nil, 1},
		})
		writeUsersSetupExpects(exporter, &writeUsersExpect{
			Users:            writeUsersExpectProp{nil, 1},
			Roles:            writeUsersExpectProp{nil, 1},
			LdapRoleMappings: writeUsersExpectProp{nil, 1},
			SamlRoleMappings: writeUsersExpectProp{err, 1},
		})
		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if fetch LDAP servers fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to read LDAP servers")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers:            fetchUsersExpectProp{nil, 1},
			GetRoles:            fetchUsersExpectProp{nil, 1},
			GetLdapRoleMappings: fetchUsersExpectProp{nil, 1},
			GetSamlRoleMappings: fetchUsersExpectProp{nil, 1},
			GetLdapServers:      fetchUsersExpectProp{err, 1},
		})
		exporter.EXPECT().
			AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()

		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if add LDAP servers file fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to write LDAP servers file")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers:            fetchUsersExpectProp{nil, 1},
			GetRoles:            fetchUsersExpectProp{nil, 1},
			GetLdapRoleMappings: fetchUsersExpectProp{nil, 1},
			GetSamlRoleMappings: fetchUsersExpectProp{nil, 1},
			GetLdapServers:      fetchUsersExpectProp{nil, 1},
		})
		writeUsersSetupExpects(exporter, &writeUsersExpect{
			Users:            writeUsersExpectProp{nil, 1},
			Roles:            writeUsersExpectProp{nil, 1},
			LdapRoleMappings: writeUsersExpectProp{nil, 1},
			SamlRoleMappings: writeUsersExpectProp{nil, 1},
			LdapServers:      writeUsersExpectProp{err, 1},
		})
		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if fetch SAML servers fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to read SAML servers")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers:                 fetchUsersExpectProp{nil, 1},
			GetRoles:                 fetchUsersExpectProp{nil, 1},
			GetLdapRoleMappings:      fetchUsersExpectProp{nil, 1},
			GetSamlRoleMappings:      fetchUsersExpectProp{nil, 1},
			GetLdapServers:           fetchUsersExpectProp{nil, 1},
			GetSamlIdentityProviders: fetchUsersExpectProp{err, 1},
		})
		exporter.EXPECT().
			AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()

		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("fails if add SAML servers file fail", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		err := fmt.Errorf("failed to write SAML servers file")
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers:                 fetchUsersExpectProp{nil, 1},
			GetRoles:                 fetchUsersExpectProp{nil, 1},
			GetLdapRoleMappings:      fetchUsersExpectProp{nil, 1},
			GetSamlRoleMappings:      fetchUsersExpectProp{nil, 1},
			GetLdapServers:           fetchUsersExpectProp{nil, 1},
			GetSamlIdentityProviders: fetchUsersExpectProp{nil, 1},
		})
		writeUsersSetupExpects(exporter, &writeUsersExpect{
			Users:                 writeUsersExpectProp{nil, 1},
			Roles:                 writeUsersExpectProp{nil, 1},
			LdapRoleMappings:      writeUsersExpectProp{nil, 1},
			SamlRoleMappings:      writeUsersExpectProp{nil, 1},
			LdapServers:           writeUsersExpectProp{nil, 1},
			SamlIdentityProviders: writeUsersExpectProp{err, 1},
		})
		result := fetchUsersData(client, exporter)

		assert.ErrorIs(t, result, err)
	})

	t.Run("succeeds if all fetch and add file succeed", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		fetchUsersSetupExpects(client, &fetchUsersExpect{
			GetUsers:                 fetchUsersExpectProp{nil, 1},
			GetRoles:                 fetchUsersExpectProp{nil, 1},
			GetLdapRoleMappings:      fetchUsersExpectProp{nil, 1},
			GetSamlRoleMappings:      fetchUsersExpectProp{nil, 1},
			GetLdapServers:           fetchUsersExpectProp{nil, 1},
			GetSamlIdentityProviders: fetchUsersExpectProp{nil, 1},
		})
		exporter.EXPECT().
			AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()

		result := fetchUsersData(client, exporter)

		assert.NoError(t, result)
	})
}
