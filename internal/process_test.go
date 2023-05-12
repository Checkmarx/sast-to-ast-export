package internal

import (
	"encoding/xml"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/persistence/installation"
	mock_preset_interfaces "github.com/checkmarxDev/ast-sast-export/test/mocks/app/preset"
	mock_integration_soap "github.com/checkmarxDev/ast-sast-export/test/mocks/integration/soap"

	"github.com/checkmarxDev/ast-sast-export/internal/app/export"
	"github.com/checkmarxDev/ast-sast-export/internal/app/metadata"
	"github.com/checkmarxDev/ast-sast-export/internal/app/querymapping"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/rest"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	mock_interfaces_query_common "github.com/checkmarxDev/ast-sast-export/test/mocks/app/ast_query"
	mock_interfaces "github.com/checkmarxDev/ast-sast-export/test/mocks/app/ast_query_mapping"
	mock_app_export "github.com/checkmarxDev/ast-sast-export/test/mocks/app/export"
	mock_app_metadata "github.com/checkmarxDev/ast-sast-export/test/mocks/app/metadata"
	mock_integration_rest "github.com/checkmarxDev/ast-sast-export/test/mocks/integration/rest"
	"github.com/golang-jwt/jwt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	TeamName   = "TestTeam"
	projectIDs = "1,2"
)

type validatePermissionTest struct {
	JwtClaims     jwt.MapClaims
	ExportOptions []string
	ExpectErr     bool
	Message       string
}

type mockExpectProps struct {
	ReturnError error
	RunCount    int
}

type usersExpect struct {
	Users            mockExpectProps
	Teams            mockExpectProps
	Roles            mockExpectProps
	LdapRoleMappings mockExpectProps
	SamlRoleMappings mockExpectProps
	LdapServers      mockExpectProps
	SamlServers      mockExpectProps
}

type teamsExpect struct {
	Teams            mockExpectProps
	LdapTeamMappings mockExpectProps
	SamlTeamMappings mockExpectProps
	LdapServers      mockExpectProps
	SamlServers      mockExpectProps
}

func fetchUsersSetupExpects(client *mock_integration_rest.MockClient, expect *usersExpect) {
	client.EXPECT().
		GetUsers().
		Return([]*rest.User{}, expect.Users.ReturnError).
		MinTimes(expect.Users.RunCount).
		MaxTimes(expect.Users.RunCount)
	client.EXPECT().
		GetTeams().
		Return([]*rest.Team{}, expect.Teams.ReturnError).
		MinTimes(expect.Teams.RunCount).
		MaxTimes(expect.Teams.RunCount)
	client.EXPECT().
		GetRoles().
		Return([]byte{}, expect.Roles.ReturnError).
		MinTimes(expect.Roles.RunCount).
		MaxTimes(expect.Roles.RunCount)
	client.EXPECT().
		GetLdapRoleMappings().
		Return([]byte{}, expect.LdapRoleMappings.ReturnError).
		MinTimes(expect.LdapRoleMappings.RunCount).
		MaxTimes(expect.LdapRoleMappings.RunCount)
	client.EXPECT().
		GetSamlRoleMappings().
		Return([]byte{}, expect.SamlRoleMappings.ReturnError).
		MinTimes(expect.SamlRoleMappings.RunCount).
		MaxTimes(expect.SamlRoleMappings.RunCount)
	client.EXPECT().
		GetLdapServers().
		Return([]byte{}, expect.LdapServers.ReturnError).
		MinTimes(expect.LdapServers.RunCount).
		MaxTimes(expect.LdapServers.RunCount)
	client.EXPECT().
		GetSamlIdentityProviders().
		Return([]byte{}, expect.SamlServers.ReturnError).
		MinTimes(expect.SamlServers.RunCount).
		MaxTimes(expect.SamlServers.RunCount)
}

func fetchTeamsSetupExpects(client *mock_integration_rest.MockClient, expect *teamsExpect) {
	client.EXPECT().
		GetTeams().
		Return([]*rest.Team{}, expect.Teams.ReturnError).
		MinTimes(expect.Teams.RunCount).
		MaxTimes(expect.Teams.RunCount)
	client.EXPECT().
		GetLdapTeamMappings().
		Return([]byte{}, expect.LdapTeamMappings.ReturnError).
		MinTimes(expect.LdapTeamMappings.RunCount).
		MaxTimes(expect.LdapTeamMappings.RunCount)
	client.EXPECT().
		GetSamlTeamMappings().
		Return([]*rest.SamlTeamMapping{}, expect.SamlTeamMappings.ReturnError).
		MinTimes(expect.SamlTeamMappings.RunCount).
		MaxTimes(expect.SamlTeamMappings.RunCount)
	client.EXPECT().
		GetLdapServers().
		Return([]byte{}, expect.LdapServers.ReturnError).
		MinTimes(expect.LdapServers.RunCount).
		MaxTimes(expect.LdapServers.RunCount)
	client.EXPECT().
		GetSamlIdentityProviders().
		Return([]byte{}, expect.SamlServers.ReturnError).
		MinTimes(expect.SamlServers.RunCount).
		MaxTimes(expect.SamlServers.RunCount)
}

func writeUsersSetupExpects(exporter *mock_app_export.MockExporter, expect *usersExpect) {
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
			return expect.SamlServers.ReturnError
		}).
		MinTimes(expect.SamlServers.RunCount).
		MaxTimes(expect.SamlServers.RunCount)
}

func writeTeamsSetupExpects(exporter *mock_app_export.MockExporter, expect *teamsExpect) {
	exporter.EXPECT().
		AddFileWithDataSource(export.TeamsFileName, gomock.Any()).
		DoAndReturn(func(_ string, _ func() ([]byte, error)) error {
			return expect.Teams.ReturnError
		}).
		MinTimes(expect.Teams.RunCount).
		MaxTimes(expect.Teams.RunCount)
	exporter.EXPECT().
		AddFileWithDataSource(export.LdapTeamMappingsFileName, gomock.Any()).
		DoAndReturn(func(_ string, _ func() ([]byte, error)) error {
			return expect.LdapTeamMappings.ReturnError
		}).
		MinTimes(expect.LdapTeamMappings.RunCount).
		MaxTimes(expect.LdapTeamMappings.RunCount)
	exporter.EXPECT().
		AddFileWithDataSource(export.SamlTeamMappingsFileName, gomock.Any()).
		DoAndReturn(func(_ string, _ func() ([]byte, error)) error {
			return expect.SamlTeamMappings.ReturnError
		}).
		MinTimes(expect.SamlTeamMappings.RunCount).
		MaxTimes(expect.SamlTeamMappings.RunCount)
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
			return expect.SamlServers.ReturnError
		}).
		MinTimes(expect.SamlServers.RunCount).
		MaxTimes(expect.SamlServers.RunCount)
}

func TestValidatePermissions(t *testing.T) {
	tests := []validatePermissionTest{
		{jwt.MapClaims{"sast-permissions": "manage-system-settings"}, []string{}, false, "empty claims and export options"},
		{jwt.MapClaims{"sast-permissions": "manage-system-settings"}, []string{export.UsersOption}, true, "empty claims"},
		{
			jwt.MapClaims{"access-control-permissions": "manage-authentication-providers", "sast-permissions": "manage-system-settings"},
			[]string{export.TeamsOption},
			false,
			"single, correct permission",
		},
		{
			jwt.MapClaims{"access-control-permissions": "invalid", "sast-permissions": "manage-system-settings"},
			[]string{export.TeamsOption},
			true,
			"single, incorrect permission",
		},
		{
			jwt.MapClaims{"access-control-permissions": nil, "sast-permissions": "manage-system-settings"},
			[]string{export.TeamsOption},
			true,
			"single, invalid permission",
		},
		{
			jwt.MapClaims{"access-control-permissions": "manage-authentication-providers", "sast-permissions": "manage-system-settings"},
			[]string{export.UsersOption},
			true,
			"missing one permission",
		},
		{
			jwt.MapClaims{
				"access-control-permissions": []interface{}{"manage-authentication-providers", "manage-roles"},
				"sast-permissions":           "manage-system-settings",
			},
			[]string{export.UsersOption},
			false,
			"permission list with correct permissions",
		},
		{
			jwt.MapClaims{"access-control-permissions": []interface{}{"invalid", "manage-roles"}, "sast-permissions": "manage-system-settings"},
			[]string{export.UsersOption},
			true,
			"permission list with incorrect permissions",
		},
		{
			jwt.MapClaims{
				"sast-permissions":           []interface{}{"use-odata", "generate-scan-report", "view-results", "manage-system-settings"},
				"access-control-permissions": []interface{}{"manage-roles", "manage-authentication-providers"},
			},
			[]string{export.UsersOption, export.ResultsOption},
			false,
			"multiple permission lists with correct permissions",
		},
		{
			jwt.MapClaims{
				"sast-permissions":           []interface{}{"invalid", "generate-scan-report", "manage-system-settings"},
				"access-control-permissions": []interface{}{"manage-roles", "manage-authentication-providers"},
			},
			[]string{export.UsersOption, export.ResultsOption},
			true,
			"multiple permission lists with incorrect permissions",
		},
		{
			jwt.MapClaims{
				"sast-permissions":           []interface{}{"use-odata", "generate-scan-report", "view-results"},
				"access-control-permissions": []interface{}{"manage-roles", "manage-authentication-providers"},
			},
			[]string{export.UsersOption, export.ResultsOption},
			true,
			"multiple permission lists with correct permissions missing manage-system-settings",
		},
	}
	for i, e := range tests {
		test := e
		t.Run(fmt.Sprintf("#%d", i+1), func(t *testing.T) {
			err := validatePermissions(test.JwtClaims, test.ExportOptions)

			if test.ExpectErr {
				assert.Error(t, err, test.Message)
			} else {
				assert.NoError(t, err, test.Message)
			}
		})
	}
}

//nolint:funlen
func TestFetchUsersData(t *testing.T) {
	t.Run("fails if any fetch fails", func(t *testing.T) {
		usersErr := fmt.Errorf("failed to read users")
		rolesErr := fmt.Errorf("failed to read roles")
		ldapMappingsErr := fmt.Errorf("failed to read LDAP role mappings")
		samlMappingsErr := fmt.Errorf("failed to read SAML role mappings")
		ldapServersErr := fmt.Errorf("failed to read LDAP servers")
		samlServersErr := fmt.Errorf("failed to read SAML servers")
		type fetchTest struct {
			mockExpects usersExpect
			expectedErr error
		}
		tests := []fetchTest{
			{
				usersExpect{
					Users: mockExpectProps{usersErr, 1},
				},
				usersErr,
			},
			{
				usersExpect{
					Users: mockExpectProps{nil, 1},
					Teams: mockExpectProps{nil, 1},
					Roles: mockExpectProps{rolesErr, 1},
				},
				rolesErr,
			},
			{
				usersExpect{
					Users:            mockExpectProps{nil, 1},
					Teams:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{ldapMappingsErr, 1},
				},
				ldapMappingsErr,
			},
			{
				usersExpect{
					Users:            mockExpectProps{nil, 1},
					Teams:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
					SamlRoleMappings: mockExpectProps{samlMappingsErr, 1},
				},
				samlMappingsErr,
			},
			{
				usersExpect{
					Users:            mockExpectProps{nil, 1},
					Teams:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
					SamlRoleMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{ldapServersErr, 1},
				},
				ldapServersErr,
			},
			{
				usersExpect{
					Users:            mockExpectProps{nil, 1},
					Teams:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
					SamlRoleMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{nil, 1},
					SamlServers:      mockExpectProps{samlServersErr, 1},
				},
				samlServersErr,
			},
		}
		// nolint:dupl
		for _, test := range tests {
			exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
			client := mock_integration_rest.NewMockClient(gomock.NewController(t))
			fetchUsersSetupExpects(client, &test.mockExpects)
			exporter.EXPECT().
				AddFileWithDataSource(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
					_, callbackErr := callback()
					return callbackErr
				}).
				AnyTimes()
			args := &Args{}

			result := fetchUsersData(client, exporter, args)

			assert.ErrorIs(t, result, test.expectedErr)
		}
	})
	t.Run("fails if any file write fails", func(t *testing.T) {
		usersErr := fmt.Errorf("failed to write users file")
		rolesErr := fmt.Errorf("failed to write roles file")
		ldapMappingsErr := fmt.Errorf("failed to write LDAP role mappings file")
		samlMappingsErr := fmt.Errorf("failed to write SAML role mappings file")
		ldapServersErr := fmt.Errorf("failed to write LDAP servers file")
		samlServersErr := fmt.Errorf("failed to write SAML servers file")
		type writeTest struct {
			fetchMockExpects usersExpect
			writeMockExpects usersExpect
			expectedErr      error
		}
		tests := []writeTest{
			{
				fetchMockExpects: usersExpect{
					Users: mockExpectProps{nil, 1},
					Teams: mockExpectProps{nil, 1},
				},
				writeMockExpects: usersExpect{
					Users: mockExpectProps{usersErr, 1},
				},
				expectedErr: usersErr,
			},
			{
				fetchMockExpects: usersExpect{
					Users: mockExpectProps{nil, 1},
					Teams: mockExpectProps{nil, 1},
					Roles: mockExpectProps{nil, 1},
				},
				writeMockExpects: usersExpect{
					Users: mockExpectProps{nil, 1},
					Roles: mockExpectProps{rolesErr, 1},
				},
				expectedErr: rolesErr,
			},
			{
				fetchMockExpects: usersExpect{
					Users:            mockExpectProps{nil, 1},
					Teams:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
				},
				writeMockExpects: usersExpect{
					Users:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{ldapMappingsErr, 1},
				},
				expectedErr: ldapMappingsErr,
			},
			{
				fetchMockExpects: usersExpect{
					Users:            mockExpectProps{nil, 1},
					Teams:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
					SamlRoleMappings: mockExpectProps{nil, 1},
				},
				writeMockExpects: usersExpect{
					Users:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
					SamlRoleMappings: mockExpectProps{samlMappingsErr, 1},
				},
				expectedErr: samlMappingsErr,
			},
			{
				fetchMockExpects: usersExpect{
					Users:            mockExpectProps{nil, 1},
					Teams:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
					SamlRoleMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{nil, 1},
				},
				writeMockExpects: usersExpect{
					Users:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
					SamlRoleMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{ldapServersErr, 1},
				},
				expectedErr: ldapServersErr,
			},
			{
				fetchMockExpects: usersExpect{
					Users:            mockExpectProps{nil, 1},
					Teams:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
					SamlRoleMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{nil, 1},
					SamlServers:      mockExpectProps{nil, 1},
				},
				writeMockExpects: usersExpect{
					Users:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
					SamlRoleMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{nil, 1},
					SamlServers:      mockExpectProps{samlServersErr, 1},
				},
				expectedErr: samlServersErr,
			},
		}
		for _, test := range tests {
			exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
			client := mock_integration_rest.NewMockClient(gomock.NewController(t))
			args := &Args{}

			fetchUsersSetupExpects(client, &test.fetchMockExpects)
			writeUsersSetupExpects(exporter, &test.writeMockExpects)

			result := fetchUsersData(client, exporter, args)

			assert.ErrorIs(t, result, test.expectedErr)
		}
	})
	t.Run("succeeds if all fetch and add file succeed", func(t *testing.T) {
		exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
		client := mock_integration_rest.NewMockClient(gomock.NewController(t))
		fetchUsersSetupExpects(client, &usersExpect{
			Users:            mockExpectProps{nil, 1},
			Teams:            mockExpectProps{nil, 1},
			Roles:            mockExpectProps{nil, 1},
			LdapRoleMappings: mockExpectProps{nil, 1},
			SamlRoleMappings: mockExpectProps{nil, 1},
			LdapServers:      mockExpectProps{nil, 1},
			SamlServers:      mockExpectProps{nil, 1},
		})
		exporter.EXPECT().
			AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()
		args := &Args{}

		result := fetchUsersData(client, exporter, args)

		assert.NoError(t, result)
	})
}

//nolint:funlen
func TestFetchTeamsData(t *testing.T) {
	t.Run("fails if any fetch fails", func(t *testing.T) {
		teamsErr := fmt.Errorf("failed to read teams")
		ldapTeamMappingsErr := fmt.Errorf("failed to read LDAP team mappings")
		samlTeamMappingsErr := fmt.Errorf("failed to read SAML team mappings")
		ldapServersErr := fmt.Errorf("failed to read LDAP servers")
		samlServersErr := fmt.Errorf("failed to read SAML servers")
		type fetchTest struct {
			mockExpects teamsExpect
			expectedErr error
		}
		tests := []fetchTest{
			{
				teamsExpect{
					Teams: mockExpectProps{teamsErr, 1},
				},
				teamsErr,
			},
			{
				teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{ldapTeamMappingsErr, 1},
				},
				ldapTeamMappingsErr,
			},
			{
				teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{nil, 1},
					SamlTeamMappings: mockExpectProps{samlTeamMappingsErr, 1},
				},
				samlTeamMappingsErr,
			},
			{
				teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{nil, 1},
					SamlTeamMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{ldapServersErr, 1},
				},
				ldapServersErr,
			},
			{
				teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{nil, 1},
					SamlTeamMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{nil, 1},
					SamlServers:      mockExpectProps{samlServersErr, 1},
				},
				samlServersErr,
			},
		}
		// nolint:dupl
		for _, test := range tests {
			exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
			client := mock_integration_rest.NewMockClient(gomock.NewController(t))
			fetchTeamsSetupExpects(client, &test.mockExpects)
			exporter.EXPECT().
				AddFileWithDataSource(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
					_, callbackErr := callback()
					return callbackErr
				}).
				AnyTimes()
			args := &Args{}

			result := fetchTeamsData(client, exporter, args)

			assert.ErrorIs(t, result, test.expectedErr)
		}
	})
	t.Run("fails if any file write fails", func(t *testing.T) {
		teamsErr := fmt.Errorf("failed to write teams file")
		ldapTeamMappingsErr := fmt.Errorf("failed to write LDAP team mappings file")
		samlTeamMappingsErr := fmt.Errorf("failed to write SAML team mappings file")
		ldapServersErr := fmt.Errorf("failed to write LDAP servers file")
		samlServersErr := fmt.Errorf("failed to write SAML servers file")
		type writeTest struct {
			fetchMockExpects teamsExpect
			writeMockExpects teamsExpect
			expectedErr      error
		}
		tests := []writeTest{
			{
				fetchMockExpects: teamsExpect{
					Teams: mockExpectProps{nil, 1},
				},
				writeMockExpects: teamsExpect{
					Teams: mockExpectProps{teamsErr, 1},
				},
				expectedErr: teamsErr,
			},
			{
				fetchMockExpects: teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{nil, 1},
				},
				writeMockExpects: teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{ldapTeamMappingsErr, 1},
				},
				expectedErr: ldapTeamMappingsErr,
			},
			{
				fetchMockExpects: teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{nil, 1},
					SamlTeamMappings: mockExpectProps{nil, 1},
				},
				writeMockExpects: teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{nil, 1},
					SamlTeamMappings: mockExpectProps{samlTeamMappingsErr, 1},
				},
				expectedErr: samlTeamMappingsErr,
			},
			{
				fetchMockExpects: teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{nil, 1},
					SamlTeamMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{nil, 1},
				},
				writeMockExpects: teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{nil, 1},
					SamlTeamMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{ldapServersErr, 1},
				},
				expectedErr: ldapServersErr,
			},
			{
				fetchMockExpects: teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{nil, 1},
					SamlTeamMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{nil, 1},
					SamlServers:      mockExpectProps{nil, 1},
				},
				writeMockExpects: teamsExpect{
					Teams:            mockExpectProps{nil, 1},
					LdapTeamMappings: mockExpectProps{nil, 1},
					SamlTeamMappings: mockExpectProps{nil, 1},
					LdapServers:      mockExpectProps{nil, 1},
					SamlServers:      mockExpectProps{samlServersErr, 1},
				},
				expectedErr: samlServersErr,
			},
		}
		for _, test := range tests {
			exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
			client := mock_integration_rest.NewMockClient(gomock.NewController(t))
			args := &Args{}

			fetchTeamsSetupExpects(client, &test.fetchMockExpects)
			writeTeamsSetupExpects(exporter, &test.writeMockExpects)

			result := fetchTeamsData(client, exporter, args)

			assert.ErrorIs(t, result, test.expectedErr)
		}
	})
	t.Run("succeeds if all fetch and add file succeed", func(t *testing.T) {
		exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
		client := mock_integration_rest.NewMockClient(gomock.NewController(t))
		fetchTeamsSetupExpects(client, &teamsExpect{
			Teams:            mockExpectProps{nil, 1},
			LdapTeamMappings: mockExpectProps{nil, 1},
			SamlTeamMappings: mockExpectProps{nil, 1},
			LdapServers:      mockExpectProps{nil, 1},
			SamlServers:      mockExpectProps{nil, 1},
		})
		exporter.EXPECT().
			AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()
		args := &Args{}

		result := fetchTeamsData(client, exporter, args)

		assert.NoError(t, result)
	})
}

func TestGetTriagedScans(t *testing.T) {
	type projectReturn struct {
		value []rest.ProjectWithLastScanID
		err   error
	}
	type resultReturn struct {
		value []rest.TriagedScanResult
		err   error
	}
	type getTriagedScansTest struct {
		projectReturns []projectReturn
		resultReturns  map[int]resultReturn
		expectedResult []TriagedScan
		expectedErr    error
		msg            string
		projectIds     string
	}
	teamName := "TestName"
	tests := []getTriagedScansTest{
		{
			projectReturns: []projectReturn{
				{
					value: []rest.ProjectWithLastScanID{
						{ID: 1, LastScanID: 1},
						{ID: 2, LastScanID: 2},
						{ID: 3, LastScanID: 3},
					},
				},
				{},
			},
			resultReturns: map[int]resultReturn{
				1: {value: []rest.TriagedScanResult{{ID: 1}}},
				2: {value: []rest.TriagedScanResult{{ID: 2}}},
				3: {value: []rest.TriagedScanResult{{ID: 3}}},
			},
			expectedResult: []TriagedScan{
				{ProjectID: 1, ScanID: 1},
				{ProjectID: 2, ScanID: 2},
				{ProjectID: 3, ScanID: 3},
			},
			expectedErr: nil,
			msg:         "success case",
			projectIds:  "1-3",
		},
		{
			projectReturns: []projectReturn{
				{
					value: []rest.ProjectWithLastScanID{},
					err:   fmt.Errorf("failed to get projects"),
				},
			},
			resultReturns:  map[int]resultReturn{},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("error searching for results"),
			msg:            "fails if can't get first project page",
			projectIds:     "1-3",
		},
		{
			projectReturns: []projectReturn{
				{
					value: []rest.ProjectWithLastScanID{
						{ID: 1, LastScanID: 1},
					},
				},
				{
					value: []rest.ProjectWithLastScanID{},
					err:   fmt.Errorf("failed to get projects"),
				},
			},
			resultReturns: map[int]resultReturn{
				1: {value: []rest.TriagedScanResult{{ID: 1}}},
			},
			expectedResult: []TriagedScan{{ProjectID: 1, ScanID: 1}},
			expectedErr:    fmt.Errorf("error searching for results"),
			msg:            "fails if can't get second project page",
			projectIds:     "1",
		},
		{
			projectReturns: []projectReturn{
				{
					value: []rest.ProjectWithLastScanID{
						{ID: 1, LastScanID: 1},
					},
				},
				{},
			},
			resultReturns: map[int]resultReturn{
				1: {
					value: []rest.TriagedScanResult{},
					err:   fmt.Errorf("failed getting result for scanID 1"),
				},
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("failed getting result for scanID 1"),
			msg:            "fails if can't get result",
			projectIds:     "1",
		},
		{
			projectReturns: []projectReturn{
				{
					value: []rest.ProjectWithLastScanID{
						{ID: 1, LastScanID: 1},
						{ID: 2, LastScanID: 2},
					},
				},
				{},
			},
			resultReturns: map[int]resultReturn{
				1: {value: []rest.TriagedScanResult{{ID: 1}}},
				2: {
					value: []rest.TriagedScanResult{},
					err:   fmt.Errorf("failed getting result for scanID 2"),
				},
			},
			expectedResult: []TriagedScan{{ProjectID: 1, ScanID: 1}},
			expectedErr:    fmt.Errorf("failed getting result for scanID 2"),
			msg:            "fails if can't get second result",
			projectIds:     projectIDs,
		},
	}
	fromDate := "2021-9-7"
	for _, test := range tests {
		client := mock_integration_rest.NewMockClient(gomock.NewController(t))
		for i, v := range test.projectReturns { //nolint:gofmt
			client.EXPECT().
				GetProjectsWithLastScanID(gomock.Eq(fromDate), gomock.Eq(teamName), gomock.Eq(test.projectIds),
					gomock.Eq(i*resultsPageLimit), gomock.Eq(resultsPageLimit)).
				Return(&test.projectReturns[i].value, v.err).
				MinTimes(1).
				MaxTimes(1)
		}
		for k, v := range test.resultReturns {
			result := test.resultReturns[k].value
			client.EXPECT().
				GetTriagedResultsByScanID(gomock.Eq(k)).
				Return(&result, v.err). //nolint:gosec
				MinTimes(1).
				MaxTimes(1)
		}

		result, err := getTriagedScans(client, fromDate, teamName, test.projectIds)

		if test.expectedErr == nil {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, test.expectedErr.Error())
		}
		assert.Equal(t, test.expectedResult, result, test.msg)
	}
}

func TestProduceReports(t *testing.T) {
	triagedScans := []TriagedScan{
		{ProjectID: 1, ScanID: 1},
		{ProjectID: 2, ScanID: 2},
	}
	reportJobs := make(chan ReportJob, 2)

	produceReports(triagedScans, reportJobs)

	expected := []ReportJob{
		{ProjectID: 1, ScanID: 1, ReportType: rest.ScanReportTypeXML},
		{ProjectID: 2, ScanID: 2, ReportType: rest.ScanReportTypeXML},
	}
	for i := 0; i < 2; i++ {
		v := <-reportJobs
		assert.Equal(t, expected[i], v)
	}
}

func TestConsumeReports(t *testing.T) {
	report1, ioErr := os.ReadFile("../test/data/process/report1.xml")
	assert.NoError(t, ioErr)
	reportCount := 4
	reportJobs := make(chan ReportJob, reportCount)
	reportJobs <- ReportJob{ProjectID: 1, ScanID: 1, ReportType: rest.ScanReportTypeXML}
	reportJobs <- ReportJob{ProjectID: 2, ScanID: 2, ReportType: rest.ScanReportTypeXML}
	reportJobs <- ReportJob{ProjectID: 3, ScanID: 3, ReportType: rest.ScanReportTypeXML}
	reportJobs <- ReportJob{ProjectID: 4, ScanID: 4, ReportType: rest.ScanReportTypeXML}
	close(reportJobs)
	ctrl := gomock.NewController(t)
	client := mock_integration_rest.NewMockClient(ctrl)
	exporter := mock_app_export.NewMockExporter(ctrl)
	client.EXPECT().CreateScanReport(gomock.Eq(1), gomock.Eq(rest.ScanReportTypeXML), gomock.Any()).
		Return(report1, nil).
		MinTimes(1).
		MaxTimes(1)
	client.EXPECT().CreateScanReport(gomock.Eq(2), gomock.Eq(rest.ScanReportTypeXML), gomock.Any()).
		Return([]byte{}, fmt.Errorf("failed getting report #2")).
		MinTimes(1).
		MaxTimes(3)
	client.EXPECT().CreateScanReport(gomock.Eq(3), gomock.Eq(rest.ScanReportTypeXML), gomock.Any()).
		Return([]byte("3"), nil).
		MinTimes(1).
		MaxTimes(1)
	client.EXPECT().CreateScanReport(gomock.Eq(4), gomock.Eq(rest.ScanReportTypeXML), gomock.Any()).
		Return(report1, nil).
		MinTimes(1).
		MaxTimes(1)
	exporter.EXPECT().AddFile(fmt.Sprintf(scansMetadataFileName, 1), gomock.Any()).Return(nil)
	exporter.EXPECT().AddFile(gomock.Eq(fmt.Sprintf(scansFileName, 1)), report1).
		DoAndReturn(func(_ string, data []byte) error {
			assert.Equal(t, string(report1), string(data))
			return nil
		}).
		MinTimes(1).
		MaxTimes(1)
	exporter.EXPECT().AddFile(gomock.Eq(fmt.Sprintf(scansFileName, 3)), gomock.Any()).
		Return(fmt.Errorf("could not write report #3")).
		MinTimes(1).
		MaxTimes(1)
	exporter.EXPECT().AddFile(fmt.Sprintf(scansMetadataFileName, 4), gomock.Any()).Return(nil)
	exporter.EXPECT().AddFile(gomock.Eq(fmt.Sprintf(scansFileName, 4)), gomock.Any()).
		Return(nil).
		MinTimes(1).
		MaxTimes(1)
	metadataProvider := mock_app_metadata.NewMockProvider(ctrl)
	metadataRecord := &metadata.Record{
		Queries: []*metadata.RecordQuery{},
	}
	metadataProvider.EXPECT().GetMetadataRecord(gomock.Any(), gomock.Any()).Return(metadataRecord, nil).AnyTimes()
	outputCh := make(chan ReportConsumeOutput, reportCount)
	args := &Args{}

	consumeReports(client, exporter, 1, reportJobs, outputCh, 3, time.Millisecond, time.Millisecond, metadataProvider, args)

	close(outputCh)
	expected := []ReportConsumeOutput{
		{Err: nil, ProjectID: 1, ScanID: 1},
		{Err: fmt.Errorf("failed getting report #2"), ProjectID: 2, ScanID: 2},
		{Err: fmt.Errorf("EOF"), ProjectID: 3, ScanID: 3},
		{Err: nil, ProjectID: 4, ScanID: 4},
	}
	for i := 0; i < reportCount; i++ {
		out := <-outputCh
		if expected[i].Err == nil {
			assert.NoError(t, out.Err)
		} else {
			assert.EqualError(t, out.Err, expected[i].Err.Error())
		}
		assert.Equal(t, expected[i].ProjectID, out.ProjectID)
		assert.Equal(t, expected[i].ScanID, out.ScanID)
	}
}

func TestFetchResultsData(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		projectPage := []rest.ProjectWithLastScanID{
			{ID: 1, LastScanID: 1},
			{ID: 2, LastScanID: 2},
		}
		teamName := TeamName
		projectsIds := projectIDs
		ctrl := gomock.NewController(t)
		client := mock_integration_rest.NewMockClient(ctrl)
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(teamName), gomock.Eq(projectsIds),
				gomock.Eq(0), gomock.Eq(resultsPageLimit)).
			Return(&projectPage, nil).
			AnyTimes()
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(teamName), gomock.Eq(projectsIds),
				gomock.Eq(resultsPageLimit), gomock.Eq(resultsPageLimit)).
			Return(&[]rest.ProjectWithLastScanID{}, nil).
			AnyTimes()
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(1)).
			Return(&[]rest.TriagedScanResult{{ID: 1}}, nil).
			AnyTimes()
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(2)).
			Return(&[]rest.TriagedScanResult{{ID: 2}}, nil).
			AnyTimes()
		client.EXPECT().CreateScanReport(gomock.Eq(1), gomock.Eq(rest.ScanReportTypeXML), gomock.Any()).
			Return([]byte("1"), nil).
			AnyTimes()
		client.EXPECT().CreateScanReport(gomock.Eq(2), gomock.Eq(rest.ScanReportTypeXML), gomock.Any()).
			Return([]byte("2"), nil).
			AnyTimes()
		exporter := mock_app_export.NewMockExporter(ctrl)
		exporter.EXPECT().AddFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)
		args := &Args{}

		result := fetchResultsData(client, exporter, 10, 3, time.Millisecond, time.Millisecond,
			metadataProvider, teamName, projectsIds, args)

		assert.NoError(t, result)
	})
	t.Run("fails if triage scans returns error", func(t *testing.T) {
		projectPage := []rest.ProjectWithLastScanID{
			{ID: 1, LastScanID: 1},
			{ID: 2, LastScanID: 2},
		}
		teamName := TeamName
		projectsIds := projectIDs
		ctrl := gomock.NewController(t)
		client := mock_integration_rest.NewMockClient(ctrl)
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(teamName), gomock.Eq(projectsIds), gomock.Eq(0),
				gomock.Eq(resultsPageLimit)).
			Return(&projectPage, nil).
			AnyTimes()
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(teamName), gomock.Eq(projectsIds),
				gomock.Eq(resultsPageLimit), gomock.Eq(resultsPageLimit)).
			Return(&[]rest.ProjectWithLastScanID{}, nil).
			AnyTimes()
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(1)).
			Return(nil, fmt.Errorf("failed getting triaged scan")).
			AnyTimes()
		exporter := mock_app_export.NewMockExporter(ctrl)
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)
		args := &Args{}

		result := fetchResultsData(client, exporter, 10, 3, time.Millisecond,
			time.Millisecond, metadataProvider, teamName, projectsIds, args)

		assert.EqualError(t, result, "failed getting triaged scan")
	})
	t.Run("doesn't fail if some results fail to fetch", func(t *testing.T) {
		projectPage := []rest.ProjectWithLastScanID{
			{ID: 1, LastScanID: 1},
			{ID: 2, LastScanID: 2},
		}
		teamName := TeamName
		projectsIds := projectIDs
		ctrl := gomock.NewController(t)
		client := mock_integration_rest.NewMockClient(ctrl)
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(teamName), gomock.Eq(projectsIds), gomock.Eq(0),
				gomock.Eq(resultsPageLimit)).
			Return(&projectPage, nil).
			AnyTimes()
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(teamName), gomock.Eq(projectsIds),
				gomock.Eq(resultsPageLimit), gomock.Eq(resultsPageLimit)).
			Return(&[]rest.ProjectWithLastScanID{}, nil).
			AnyTimes()
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(1)).
			Return(&[]rest.TriagedScanResult{{ID: 1}}, nil).
			AnyTimes()
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(2)).
			Return(&[]rest.TriagedScanResult{{ID: 2}}, nil).
			AnyTimes()
		client.EXPECT().CreateScanReport(gomock.Eq(1), gomock.Eq(rest.ScanReportTypeXML), gomock.Any()).
			Return([]byte("1"), nil).
			AnyTimes()
		client.EXPECT().CreateScanReport(gomock.Eq(2), gomock.Eq(rest.ScanReportTypeXML), gomock.Any()).
			Return([]byte{}, fmt.Errorf("failed getting report #2")).
			AnyTimes()
		exporter := mock_app_export.NewMockExporter(ctrl)
		exporter.EXPECT().AddFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)
		args := &Args{}

		result := fetchResultsData(client, exporter, 10, 3, time.Millisecond,
			time.Millisecond, metadataProvider, teamName, projectsIds, args)

		assert.NoError(t, result)
	})
}

//nolint:funlen
func TestFetchSelectedData(t *testing.T) {
	teamName := TeamName
	projectsIds := projectIDs
	projectPage := []rest.ProjectWithLastScanID{
		{ID: 1, LastScanID: 1},
		{ID: 2, LastScanID: 2},
	}
	t.Run("export users success case", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(TeamName), gomock.Eq(projectIDs),
				gomock.Eq(0), gomock.Eq(resultsPageLimit)).
			Return(&projectPage, nil).
			AnyTimes()
		client.EXPECT().GetUsers().Return([]*rest.User{}, nil)
		client.EXPECT().GetTeams().Return([]*rest.Team{}, nil)
		exporter := mock_app_export.NewMockExporter(ctrl)
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		args := Args{
			Export:              []string{"users"},
			ProjectsActiveSince: 100,
		}
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond,
			metadataProvider, queryProvider, presetProvider)

		assert.NoError(t, result)
	})
	t.Run("export projects and presets success case", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		var preset100000 soap.GetPresetDetailsResponse
		presetList := []*rest.PresetShort{
			{ID: 1, Name: "All", OwnerName: "CxUser"},
			{ID: 9, Name: "Android", OwnerName: "CxUser"},
			{ID: 100000, Name: "New_custom_preset", OwnerName: "Custom_user"},
			{ID: 100001, Name: "New_custom_preset_2", OwnerName: "Custom_user"}, // this one should be ignored
		}
		projects := []*rest.Project{{ID: 1, Name: "test_name", IsPublic: true, TeamID: 1,
			CreatedDate: "2022-04-21T20:30:59.39+03:00", PresetID: 100000,
			Configuration: &rest.Configuration{
				CustomFields: []*rest.CustomField{{FieldName: "Creator_custom_field", FieldValue: "test"}},
			}}}
		presetXML100000, io100000Err := os.ReadFile("../test/data/presets/100000.xml")
		assert.NoError(t, io100000Err)
		unmarshal100000Err := xml.Unmarshal(presetXML100000, &preset100000)
		assert.NoError(t, unmarshal100000Err)
		exporter := mock_app_export.NewMockExporter(ctrl)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)

		client.EXPECT().GetProjects(gomock.Any(), teamName, projectsIds, 0,
			gomock.Any()).Return(projects, nil)
		client.EXPECT().GetProjects(gomock.Any(), teamName, projectsIds, gomock.Any(),
			gomock.Any()).Return([]*rest.Project{}, nil)
		client.EXPECT().GetPresets().Return(presetList, nil).Times(1)
		presetProvider.EXPECT().GetPresetDetails(100000).Return(&preset100000, nil)
		exporter.EXPECT().CreateDir(export.PresetsDirName).Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		exporter.EXPECT().AddFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		args := Args{
			Export:              []string{"presets", "projects"},
			ProjectsIds:         projectIDs,
			TeamName:            teamName,
			ProjectsActiveSince: 100,
		}
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond,
			metadataProvider, queryProvider, presetProvider)

		assert.NoError(t, result)
	})
	t.Run("export all presets success case", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		var preset1, preset9, preset100000, preset100001 soap.GetPresetDetailsResponse
		presetList := []*rest.PresetShort{
			{ID: 1, Name: "All", OwnerName: "CxUser"},
			{ID: 9, Name: "Android", OwnerName: "CxUser"},
			{ID: 100000, Name: "New_custom_preset", OwnerName: "Custom_user"},
			{ID: 100001, Name: "New_custom_preset_2", OwnerName: "Custom_user"}, // this one should not be ignored
		}
		args := Args{
			Export:              []string{"presets"},
			TeamName:            teamName,
			ProjectsActiveSince: 100,
		}
		presetXML1, io1Err := os.ReadFile("../test/data/presets/1.xml")
		assert.NoError(t, io1Err)
		unmarshal1Err := xml.Unmarshal(presetXML1, &preset1)
		assert.NoError(t, unmarshal1Err)
		presetXML9, io9Err := os.ReadFile("../test/data/presets/9.xml")
		assert.NoError(t, io9Err)
		unmarshal9Err := xml.Unmarshal(presetXML9, &preset9)
		assert.NoError(t, unmarshal9Err)
		presetXML100000, io100000Err := os.ReadFile("../test/data/presets/100000.xml")
		assert.NoError(t, io100000Err)
		unmarshal100000Err := xml.Unmarshal(presetXML100000, &preset100000)
		assert.NoError(t, unmarshal100000Err)
		presetXML100001, io100001Err := os.ReadFile("../test/data/presets/100001.xml")
		assert.NoError(t, io100001Err)
		unmarshal100001Err := xml.Unmarshal(presetXML100001, &preset100001)
		assert.NoError(t, unmarshal100001Err)
		exporter := mock_app_export.NewMockExporter(ctrl)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)

		client.EXPECT().GetPresets().Return(presetList, nil).AnyTimes()
		presetProvider.EXPECT().GetPresetDetails(1).Return(&preset1, nil)
		presetProvider.EXPECT().GetPresetDetails(9).Return(&preset9, nil)
		presetProvider.EXPECT().GetPresetDetails(100000).Return(&preset100000, nil)
		presetProvider.EXPECT().GetPresetDetails(100001).Return(&preset100001, nil)
		exporter.EXPECT().CreateDir(export.PresetsDirName).Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		exporter.EXPECT().AddFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond,
			metadataProvider, queryProvider, presetProvider)

		assert.NoError(t, result)
	})
	t.Run("export users fails if fetch or write fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)
		client.EXPECT().GetUsers().Return([]*rest.User{}, nil)
		client.EXPECT().GetTeams().Return([]*rest.Team{}, nil)
		exporter := mock_app_export.NewMockExporter(ctrl)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.UsersFileName), gomock.Any()).
			Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.RolesFileName), gomock.Any()).
			Return(fmt.Errorf("failed fetching roles"))
		args := Args{
			Export:              []string{"users"},
			ProjectsActiveSince: 100,
		}
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond,
			metadataProvider, queryProvider, presetProvider)

		assert.EqualError(t, result, "failed fetching roles")
	})
	t.Run("export users and teams success case", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)
		client.EXPECT().GetUsers().Return([]*rest.User{}, nil)
		client.EXPECT().GetTeams().Return([]*rest.Team{}, nil).Times(2)
		client.EXPECT().GetSamlTeamMappings().Return([]*rest.SamlTeamMapping{}, nil)
		exporter := mock_app_export.NewMockExporter(ctrl)
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		args := Args{
			Export:              []string{"users", "teams"},
			ProjectsActiveSince: 100,
		}
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond,
			metadataProvider, queryProvider, presetProvider)

		assert.NoError(t, result)
	})
	t.Run("export users and teams fail if fetch or write fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)
		client.EXPECT().GetUsers().Return([]*rest.User{}, nil)
		client.EXPECT().GetTeams().Return([]*rest.Team{}, nil).Times(2)
		exporter := mock_app_export.NewMockExporter(ctrl)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.UsersFileName), gomock.Any()).
			Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.RolesFileName), gomock.Any()).
			Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.LdapRoleMappingsFileName), gomock.Any()).
			Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.SamlRoleMappingsFileName), gomock.Any()).
			Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.LdapServersFileName), gomock.Any()).
			Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.SamlIdpFileName), gomock.Any()).
			Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.TeamsFileName), gomock.Any()).
			Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.LdapTeamMappingsFileName), gomock.Any()).
			Return(fmt.Errorf("failed fetching LDAP team mappings"))
		args := Args{
			Export:              []string{"users", "teams"},
			ProjectsActiveSince: 100,
		}
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond,
			metadataProvider, queryProvider, presetProvider)

		assert.EqualError(t, result, "failed fetching LDAP team mappings")
	})
	t.Run("export users, teams and results success case", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)
		projectPage := []rest.ProjectWithLastScanID{
			{ID: 1, LastScanID: 1},
			{ID: 2, LastScanID: 2},
		}
		client.EXPECT().GetUsers().Return([]*rest.User{}, nil)
		client.EXPECT().GetTeams().Return([]*rest.Team{}, nil).Times(2)
		client.EXPECT().GetSamlTeamMappings().Return([]*rest.SamlTeamMapping{}, nil)
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(0),
				gomock.Any()).
			Return(&projectPage, nil)
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any()).
			Return(&[]rest.ProjectWithLastScanID{}, nil)
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(1)).
			Return(&[]rest.TriagedScanResult{{ID: 1}}, nil)
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(2)).
			Return(&[]rest.TriagedScanResult{{ID: 2}}, nil)
		client.EXPECT().
			CreateScanReport(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]byte("test"), nil).
			AnyTimes()
		exporter := mock_app_export.NewMockExporter(ctrl)
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		exporter.EXPECT().AddFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		args := Args{
			Export:              []string{export.UsersOption, export.TeamsOption, export.ResultsOption},
			ProjectsActiveSince: 100,
		}
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond,
			metadataProvider, queryProvider, presetProvider)

		assert.NoError(t, result)
	})
	t.Run("export users, teams and results fails if result processing fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)
		client.EXPECT().GetUsers().Return([]*rest.User{}, nil)
		client.EXPECT().GetTeams().Return([]*rest.Team{}, nil).Times(2)
		client.EXPECT().GetSamlTeamMappings().Return([]*rest.SamlTeamMapping{}, nil)
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(0),
				gomock.Any()).
			Return(&[]rest.ProjectWithLastScanID{}, fmt.Errorf("failed fetching projects"))
		exporter := mock_app_export.NewMockExporter(ctrl)
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		args := Args{
			Export:              []string{export.UsersOption, export.TeamsOption, export.ResultsOption},
			ProjectsActiveSince: 100,
		}
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond,
			metadataProvider, queryProvider, presetProvider)

		assert.EqualError(t, result, "error searching for results")
	})
	t.Run("empty export if no export options selected", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)
		exporter := mock_app_export.NewMockExporter(ctrl)
		args := Args{
			Export:              []string{},
			ProjectsActiveSince: 100,
		}
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond,
			metadataProvider, queryProvider, presetProvider)

		assert.NoError(t, result)
	})
	t.Run("empty export if export options are invalid", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)
		exporter := mock_app_export.NewMockExporter(ctrl)
		args := Args{
			Export:              []string{"test1", "test2"},
			ProjectsActiveSince: 100,
		}
		metadataProvider := mock_app_metadata.NewMockProvider(ctrl)

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond,
			metadataProvider, queryProvider, presetProvider)

		assert.NoError(t, result)
	})
}

func TestExportResultsToFile(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		args := Args{
			Debug:       false,
			ProductName: "test",
			OutputPath:  "/path/to/output",
		}
		ctrl := gomock.NewController(t)
		exporter := mock_app_export.NewMockExporter(ctrl)
		exporter.EXPECT().GetTmpDir().Return("/path/to/tmp/folder").MinTimes(1).MaxTimes(1)
		exporter.EXPECT().CreateExportPackage(gomock.Eq(args.ProductName), gomock.Eq(args.OutputPath)).
			Return("/path/to/output/export.zip", "/path/to/output/key.txt", nil).
			MinTimes(1).
			MaxTimes(1)

		fileName, err := exportResultsToFile(&args, exporter)

		assert.NoError(t, err)
		assert.Equal(t, "/path/to/output/export.zip", fileName)
	})
	t.Run("fails if export package creation fails", func(t *testing.T) {
		args := Args{
			Debug:       false,
			ProductName: "test",
			OutputPath:  "/path/to/output",
		}
		ctrl := gomock.NewController(t)
		exporter := mock_app_export.NewMockExporter(ctrl)
		exporter.EXPECT().GetTmpDir().Return("/path/to/tmp/folder").MinTimes(1).MaxTimes(1)
		exporter.EXPECT().CreateExportPackage(gomock.Eq(args.ProductName), gomock.Eq(args.OutputPath)).
			Return("", "", fmt.Errorf("failed creating export package")).
			MinTimes(1).
			MaxTimes(1)

		fileName, err := exportResultsToFile(&args, exporter)

		assert.EqualError(t, err, "failed creating export package")
		assert.Equal(t, "", fileName)
	})
}

func TestFetchProjects(t *testing.T) {
	teamName := TeamName
	projectsIds := projectIDs
	t.Run("fetch projects successfully", func(t *testing.T) {
		projects := []*rest.Project{{ID: 1, Name: "test_name", IsPublic: true, TeamID: 1,
			CreatedDate: "2022-04-21T20:30:59.39+03:00",
			Configuration: &rest.Configuration{
				CustomFields: []*rest.CustomField{{FieldName: "Creator_custom_field", FieldValue: "test"}},
			}}}
		exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
		client := mock_integration_rest.NewMockClient(gomock.NewController(t))
		client.EXPECT().GetProjects(gomock.Any(), teamName, projectsIds, 0,
			gomock.Any()).Return(projects, nil)
		client.EXPECT().GetProjects(gomock.Any(), teamName, projectsIds, gomock.Any(),
			gomock.Any()).Return([]*rest.Project{}, nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()

		projectsList, errProjects := fetchProjectsData(client, exporter, 10, teamName, projectsIds)

		assert.NoError(t, errProjects)
		assert.Equal(t, projects, projectsList)
	})

	t.Run("fetch projects with error", func(t *testing.T) {
		exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
		client := mock_integration_rest.NewMockClient(gomock.NewController(t))
		client.EXPECT().GetProjects(gomock.Any(), teamName, projectsIds, 0,
			gomock.Any()).Return([]*rest.Project{}, fmt.Errorf("failed fetching project")).Times(1)

		_, err := fetchProjectsData(client, exporter, 10, teamName, projectsIds)

		assert.EqualError(t, err, "failed getting projects: failed fetching project")
	})

	t.Run("fetch many pages", func(t *testing.T) {
		projectsIds = "1-4"
		projectsFirst := []*rest.Project{{ID: 1, Name: "test_name", IsPublic: true, TeamID: 1,
			CreatedDate: "2022-04-21T20:30:59.39+03:00",
			Configuration: &rest.Configuration{
				CustomFields: []*rest.CustomField{{FieldName: "Creator_custom_field", FieldValue: "test"}},
			}}}
		projectsSecond := []*rest.Project{{ID: 4, Name: "test_name 4", IsPublic: true, TeamID: 1,
			CreatedDate: "2022-04-22T20:30:59.39+03:00",
			Configuration: &rest.Configuration{
				CustomFields: []*rest.CustomField{{FieldName: "Creator_custom_field", FieldValue: "test 4"}},
			}}}
		expectedList := []*rest.Project{projectsFirst[0], projectsSecond[0]}
		exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
		client := mock_integration_rest.NewMockClient(gomock.NewController(t))
		client.EXPECT().GetProjects(gomock.Any(), teamName, projectsIds, 0,
			gomock.Any()).Return(projectsFirst, nil)
		client.EXPECT().GetProjects(gomock.Any(), teamName, projectsIds, gomock.Any(),
			gomock.Any()).Return(projectsSecond, nil)
		client.EXPECT().GetProjects(gomock.Any(), teamName, projectsIds, gomock.Any(),
			gomock.Any()).Return([]*rest.Project{}, nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).
			AnyTimes()

		projectsList, errProjects := fetchProjectsData(client, exporter, 10, teamName, projectsIds)

		assert.NoError(t, errProjects)
		assert.Equal(t, expectedList, projectsList)
	})
}

func TestCustomQueries(t *testing.T) {
	t.Run("fetch custom queries", func(t *testing.T) {
		var customQueriesObj soap.GetQueryCollectionResponse
		ctrl := gomock.NewController(t)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
		customQueries, ioCustomErr := os.ReadFile("../test/data/queries/custom_queries.xml")
		assert.NoError(t, ioCustomErr)
		_ = xml.Unmarshal(customQueries, &customQueriesObj)
		queryProvider.EXPECT().GetCustomQueriesList().Return(&customQueriesObj, nil).Times(1)
		exporter.EXPECT().AddFile(export.QueriesFileName, gomock.Any()).Return(nil)

		result := fetchQueriesData(queryProvider, exporter)

		assert.NoError(t, result)
	})
}

func TestPresets(t *testing.T) {
	presetList := []*rest.PresetShort{
		{ID: 1, Name: "All", OwnerName: "CxUser"},
		{ID: 9, Name: "Android", OwnerName: "CxUser"},
		{ID: 100000, Name: "New_custom_preset", OwnerName: "Custom_user"},
	}

	t.Run("fetch custom and default presets successfully", func(t *testing.T) {
		var preset100000 soap.GetPresetDetailsResponse
		var preset1 soap.GetPresetDetailsResponse
		var preset9 soap.GetPresetDetailsResponse
		ctrl := gomock.NewController(t)
		exporter := mock_app_export.NewMockExporter(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)
		presetXML100000, io100000Err := os.ReadFile("../test/data/presets/100000.xml")
		assert.NoError(t, io100000Err)
		presetXML1, io1Err := os.ReadFile("../test/data/presets/1.xml")
		assert.NoError(t, io1Err)
		presetXML9, io9Err := os.ReadFile("../test/data/presets/9.xml")
		assert.NoError(t, io9Err)
		_ = xml.Unmarshal(presetXML100000, &preset100000)
		_ = xml.Unmarshal(presetXML1, &preset1)
		_ = xml.Unmarshal(presetXML9, &preset9)
		client.EXPECT().GetPresets().Return(presetList, nil).Times(1)
		presetProvider.EXPECT().GetPresetDetails(100000).Return(&preset100000, nil)
		presetProvider.EXPECT().GetPresetDetails(1).Return(&preset1, nil)
		presetProvider.EXPECT().GetPresetDetails(9).Return(&preset9, nil)
		exporter.EXPECT().CreateDir(export.PresetsDirName).Return(nil)
		exporter.EXPECT().AddFileWithDataSource(export.PresetsFileName, gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).AnyTimes()
		exporter.EXPECT().AddFile(path.Join(export.PresetsDirName, "100000.xml"), gomock.Any()).Return(nil)
		exporter.EXPECT().AddFile(path.Join(export.PresetsDirName, "1.xml"), gomock.Any()).Return(nil)
		exporter.EXPECT().AddFile(path.Join(export.PresetsDirName, "9.xml"), gomock.Any()).Return(nil)

		err := fetchPresetsData(client, presetProvider, exporter, nil, "")

		assert.NoError(t, err)
	})
	t.Run("error with fetching presets in REST client", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		exporter := mock_app_export.NewMockExporter(ctrl)
		presetProvider := mock_preset_interfaces.NewMockPresetProvider(ctrl)
		client := mock_integration_rest.NewMockClient(ctrl)
		client.EXPECT().GetPresets().Return(nil, fmt.Errorf("failed getting preset list")).Times(1)

		err := fetchPresetsData(client, presetProvider, exporter, nil, "")

		assert.EqualError(t, err, "error with getting preset list: failed getting preset list")
		assert.Error(t, err)
	})
}

func TestAddQueryMappingFile(t *testing.T) {
	t.Run("test add query mapping file", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		exporter := mock_app_export.NewMockExporter(ctrl)
		queryMappingProvider := mock_interfaces.NewMockQueryMappingRepo(ctrl)
		exporter.EXPECT().AddFileWithDataSource(destQueryMappingFile, gomock.Any()).
			DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
				_, callbackErr := callback()
				return callbackErr
			}).Times(1)
		queryMappingProvider.EXPECT().GetMapping().Return([]querymapping.QueryMap{}).Times(1)

		err := addQueryMappingFile(queryMappingProvider, exporter)
		assert.NoError(t, err)
	})
}

func TestAddCustomQueryIDs(t *testing.T) {
	t.Run("test add custom query to mapping", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		queryMappingProvider := mock_interfaces.NewMockQueryMappingRepo(ctrl)
		queryProvider := mock_interfaces_query_common.NewMockASTQueryProvider(ctrl)
		queryProvider.EXPECT().GetCustomQueriesList().Return(&soap.GetQueryCollectionResponse{
			GetQueryCollectionResult: soap.GetQueryCollectionResult{
				QueryGroups: soap.QueryGroups{
					CxWSQueryGroup: []soap.CxWSQueryGroup{
						{
							Name:         "Test_group",
							LanguageName: "Go",
							Queries: soap.Queries{
								CxWSQuery: []soap.CxWSQuery{
									{QueryID: 1, Name: "Test_query"},
								},
							},
						},
					},
				},
			},
		}, nil).Times(1)
		queryMappingProvider.EXPECT().AddQueryMapping("Go", "Test_query", "Test_group", "1").Return(nil).Times(1)

		err := addCustomQueryIDs(queryProvider, queryMappingProvider)
		assert.NoError(t, err)
	})
}

func TestFetchInstallationData(t *testing.T) {
	soapResponseSuccess := &soap.GetInstallationSettingsResponse{
		GetInstallationSettingsResult: soap.GetInstallationSettingsResult{
			IsSuccesfull: "true",
			InstallationSettingsList: soap.InstallationSettingsList{
				InstallationSetting: []*soap.InstallationSetting{
					{
						Name:    "Checkmarx Engine Service",
						Version: "9.3.4.1111",
						Hotfix:  "Hotfix",
					},
					{
						Name:    "Checkmarx Queries Pack",
						Version: "9.3.4.5111",
						Hotfix:  "Hotfix",
					},
				},
			},
		},
	}
	t.Run("test add installation version", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
		soapClientMock.EXPECT().GetInstallationSettings().Return(soapResponseSuccess, nil)
		exporter := mock_app_export.NewMockExporter(gomock.NewController(t))
		exporter.EXPECT().
			AddFileWithDataSource(export.InstallationFileName, gomock.Any()).
			DoAndReturn(func(_ string, _ func() ([]byte, error)) error {
				return nil
			})

		instance := installation.NewRepo(soapClientMock)

		err := fetchInstallationData(instance, exporter)
		assert.NoError(t, err)
	})
}
