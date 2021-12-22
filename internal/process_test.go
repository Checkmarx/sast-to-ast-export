package internal

import (
	"fmt"
	"testing"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/sast"
	export2 "github.com/checkmarxDev/ast-sast-export/internal/test/mocks/export"
	sast2 "github.com/checkmarxDev/ast-sast-export/internal/test/mocks/sast"
	"github.com/golang/mock/gomock"

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

type mockExpectProps struct {
	ReturnError error
	RunCount    int
}

type usersExpect struct {
	Users            mockExpectProps
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

func fetchUsersSetupExpects(client *sast2.MockClient, expect *usersExpect) {
	client.EXPECT().
		GetUsers().
		Return([]byte{}, expect.Users.ReturnError).
		MinTimes(expect.Users.RunCount).
		MaxTimes(expect.Users.RunCount)
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

func fetchTeamsSetupExpects(client *sast2.MockClient, expect *teamsExpect) {
	client.EXPECT().
		GetTeams().
		Return([]byte{}, expect.Teams.ReturnError).
		MinTimes(expect.Teams.RunCount).
		MaxTimes(expect.Teams.RunCount)
	client.EXPECT().
		GetLdapTeamMappings().
		Return([]byte{}, expect.LdapTeamMappings.ReturnError).
		MinTimes(expect.LdapTeamMappings.RunCount).
		MaxTimes(expect.LdapTeamMappings.RunCount)
	client.EXPECT().
		GetSamlTeamMappings().
		Return([]byte{}, expect.SamlTeamMappings.ReturnError).
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

func writeUsersSetupExpects(exporter *export2.MockExporter, expect *usersExpect) {
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

func writeTeamsSetupExpects(exporter *export2.MockExporter, expect *teamsExpect) {
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
					Roles: mockExpectProps{rolesErr, 1},
				},
				rolesErr,
			},
			{
				usersExpect{
					Users:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{ldapMappingsErr, 1},
				},
				ldapMappingsErr,
			},
			{
				usersExpect{
					Users:            mockExpectProps{nil, 1},
					Roles:            mockExpectProps{nil, 1},
					LdapRoleMappings: mockExpectProps{nil, 1},
					SamlRoleMappings: mockExpectProps{samlMappingsErr, 1},
				},
				samlMappingsErr,
			},
			{
				usersExpect{
					Users:            mockExpectProps{nil, 1},
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
			exporter := export2.NewMockExporter(gomock.NewController(t))
			client := sast2.NewMockClient(gomock.NewController(t))
			fetchUsersSetupExpects(client, &test.mockExpects)
			exporter.EXPECT().
				AddFileWithDataSource(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
					_, callbackErr := callback()
					return callbackErr
				}).
				AnyTimes()

			result := fetchUsersData(client, exporter)

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
				},
				writeMockExpects: usersExpect{
					Users: mockExpectProps{usersErr, 1},
				},
				expectedErr: usersErr,
			},
			{
				fetchMockExpects: usersExpect{
					Users: mockExpectProps{nil, 1},
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
			exporter := export2.NewMockExporter(gomock.NewController(t))
			client := sast2.NewMockClient(gomock.NewController(t))

			fetchUsersSetupExpects(client, &test.fetchMockExpects)
			writeUsersSetupExpects(exporter, &test.writeMockExpects)

			result := fetchUsersData(client, exporter)

			assert.ErrorIs(t, result, test.expectedErr)
		}
	})
	t.Run("succeeds if all fetch and add file succeed", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
		fetchUsersSetupExpects(client, &usersExpect{
			Users:            mockExpectProps{nil, 1},
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

		result := fetchUsersData(client, exporter)

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
			exporter := export2.NewMockExporter(gomock.NewController(t))
			client := sast2.NewMockClient(gomock.NewController(t))
			fetchTeamsSetupExpects(client, &test.mockExpects)
			exporter.EXPECT().
				AddFileWithDataSource(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ string, callback func() ([]byte, error)) error {
					_, callbackErr := callback()
					return callbackErr
				}).
				AnyTimes()

			result := fetchTeamsData(client, exporter)

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
			exporter := export2.NewMockExporter(gomock.NewController(t))
			client := sast2.NewMockClient(gomock.NewController(t))

			fetchTeamsSetupExpects(client, &test.fetchMockExpects)
			writeTeamsSetupExpects(exporter, &test.writeMockExpects)

			result := fetchTeamsData(client, exporter)

			assert.ErrorIs(t, result, test.expectedErr)
		}
	})
	t.Run("succeeds if all fetch and add file succeed", func(t *testing.T) {
		exporter := export2.NewMockExporter(gomock.NewController(t))
		client := sast2.NewMockClient(gomock.NewController(t))
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

		result := fetchTeamsData(client, exporter)

		assert.NoError(t, result)
	})
}

func TestGetTriagedScans(t *testing.T) {
	type projectReturn struct {
		value []sast.ProjectWithLastScanID
		err   error
	}
	type resultReturn struct {
		value []sast.TriagedScanResult
		err   error
	}
	type getTriagedScansTest struct {
		projectReturns []projectReturn
		resultReturns  map[int]resultReturn
		expectedResult []TriagedScan
		expectedErr    error
		msg            string
	}
	tests := []getTriagedScansTest{
		{
			projectReturns: []projectReturn{
				{
					value: []sast.ProjectWithLastScanID{
						{ID: 1, LastScanID: 1},
						{ID: 2, LastScanID: 2},
						{ID: 3, LastScanID: 3},
					},
				},
				{},
			},
			resultReturns: map[int]resultReturn{
				1: {value: []sast.TriagedScanResult{{ID: 1}}},
				2: {value: []sast.TriagedScanResult{{ID: 2}}},
				3: {value: []sast.TriagedScanResult{{ID: 3}}},
			},
			expectedResult: []TriagedScan{
				{ProjectID: 1, ScanID: 1},
				{ProjectID: 2, ScanID: 2},
				{ProjectID: 3, ScanID: 3},
			},
			expectedErr: nil,
			msg:         "success case",
		},
		{
			projectReturns: []projectReturn{
				{
					value: []sast.ProjectWithLastScanID{},
					err:   fmt.Errorf("failed to get projects"),
				},
			},
			resultReturns:  map[int]resultReturn{},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("error searching for results"),
			msg:            "fails if can't get first project page",
		},
		{
			projectReturns: []projectReturn{
				{
					value: []sast.ProjectWithLastScanID{
						{ID: 1, LastScanID: 1},
					},
				},
				{
					value: []sast.ProjectWithLastScanID{},
					err:   fmt.Errorf("failed to get projects"),
				},
			},
			resultReturns: map[int]resultReturn{
				1: {value: []sast.TriagedScanResult{{ID: 1}}},
			},
			expectedResult: []TriagedScan{{ProjectID: 1, ScanID: 1}},
			expectedErr:    fmt.Errorf("error searching for results"),
			msg:            "fails if can't get second project page",
		},
		{
			projectReturns: []projectReturn{
				{
					value: []sast.ProjectWithLastScanID{
						{ID: 1, LastScanID: 1},
					},
				},
				{},
			},
			resultReturns: map[int]resultReturn{
				1: {
					value: []sast.TriagedScanResult{},
					err:   fmt.Errorf("failed getting result for scanID 1"),
				},
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("failed getting result for scanID 1"),
			msg:            "fails if can't get result",
		},
		{
			projectReturns: []projectReturn{
				{
					value: []sast.ProjectWithLastScanID{
						{ID: 1, LastScanID: 1},
						{ID: 2, LastScanID: 2},
					},
				},
				{},
			},
			resultReturns: map[int]resultReturn{
				1: {value: []sast.TriagedScanResult{{ID: 1}}},
				2: {
					value: []sast.TriagedScanResult{},
					err:   fmt.Errorf("failed getting result for scanID 2"),
				},
			},
			expectedResult: []TriagedScan{{ProjectID: 1, ScanID: 1}},
			expectedErr:    fmt.Errorf("failed getting result for scanID 2"),
			msg:            "fails if can't get second result",
		},
	}
	fromDate := "2021-9-7"
	for _, test := range tests {
		client := sast2.NewMockClient(gomock.NewController(t))
		for i, v := range test.projectReturns { //nolint:gofmt
			client.EXPECT().
				GetProjectsWithLastScanID(gomock.Eq(fromDate), gomock.Eq(i*resultsPageLimit), gomock.Eq(resultsPageLimit)).
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

		result, err := getTriagedScans(client, fromDate)

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
		{ProjectID: 1, ScanID: 1, ReportType: sast.ScanReportTypeXML},
		{ProjectID: 2, ScanID: 2, ReportType: sast.ScanReportTypeXML},
	}
	for i := 0; i < 2; i++ {
		v := <-reportJobs
		assert.Equal(t, expected[i], v)
	}
}

func TestConsumeReports(t *testing.T) {
	reportCount := 4
	reportJobs := make(chan ReportJob, reportCount)
	reportJobs <- ReportJob{ProjectID: 1, ScanID: 1, ReportType: sast.ScanReportTypeXML}
	reportJobs <- ReportJob{ProjectID: 2, ScanID: 2, ReportType: sast.ScanReportTypeXML}
	reportJobs <- ReportJob{ProjectID: 3, ScanID: 3, ReportType: sast.ScanReportTypeXML}
	reportJobs <- ReportJob{ProjectID: 4, ScanID: 4, ReportType: sast.ScanReportTypeXML}
	close(reportJobs)
	client := sast2.NewMockClient(gomock.NewController(t))
	exporter := export2.NewMockExporter(gomock.NewController(t))
	client.EXPECT().CreateScanReport(gomock.Eq(1), gomock.Eq(sast.ScanReportTypeXML), gomock.Any()).
		Return([]byte("1"), nil).
		MinTimes(1).
		MaxTimes(1)
	client.EXPECT().CreateScanReport(gomock.Eq(2), gomock.Eq(sast.ScanReportTypeXML), gomock.Any()).
		Return([]byte{}, fmt.Errorf("failed getting report #2")).
		MinTimes(1).
		MaxTimes(3)
	client.EXPECT().CreateScanReport(gomock.Eq(3), gomock.Eq(sast.ScanReportTypeXML), gomock.Any()).
		Return([]byte("3"), nil).
		MinTimes(1).
		MaxTimes(1)

	client.EXPECT().CreateScanReport(gomock.Eq(4), gomock.Eq(sast.ScanReportTypeXML), gomock.Any()).
		Return([]byte("4"), nil).
		MinTimes(1).
		MaxTimes(1)
	exporter.EXPECT().AddFile(gomock.Eq(fmt.Sprintf(scansFileName, 1)), gomock.Any()).
		Return(nil).
		MinTimes(1).
		MaxTimes(1)
	exporter.EXPECT().AddFile(gomock.Eq(fmt.Sprintf(scansFileName, 3)), gomock.Any()).
		Return(fmt.Errorf("could not write report #3")).
		MinTimes(1).
		MaxTimes(1)
	exporter.EXPECT().AddFile(gomock.Eq(fmt.Sprintf(scansFileName, 4)), gomock.Any()).
		Return(nil).
		MinTimes(1).
		MaxTimes(1)
	outputCh := make(chan ReportConsumeOutput, reportCount)

	consumeReports(client, exporter, 1, reportJobs, outputCh, 3, time.Millisecond, time.Millisecond)

	close(outputCh)
	expected := []ReportConsumeOutput{
		{Err: nil, ProjectID: 1, ScanID: 1},
		{Err: fmt.Errorf("failed getting report #2"), ProjectID: 2, ScanID: 2},
		{Err: fmt.Errorf("could not write report #3"), ProjectID: 3, ScanID: 3},
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
		projectPage := []sast.ProjectWithLastScanID{
			{ID: 1, LastScanID: 1},
			{ID: 2, LastScanID: 2},
		}
		client := sast2.NewMockClient(gomock.NewController(t))
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(0), gomock.Eq(resultsPageLimit)).
			Return(&projectPage, nil).
			AnyTimes()
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(resultsPageLimit), gomock.Eq(resultsPageLimit)).
			Return(&[]sast.ProjectWithLastScanID{}, nil).
			AnyTimes()
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(1)).
			Return(&[]sast.TriagedScanResult{{ID: 1}}, nil).
			AnyTimes()
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(2)).
			Return(&[]sast.TriagedScanResult{{ID: 2}}, nil).
			AnyTimes()
		client.EXPECT().CreateScanReport(gomock.Eq(1), gomock.Eq(sast.ScanReportTypeXML), gomock.Any()).
			Return([]byte("1"), nil).
			AnyTimes()
		client.EXPECT().CreateScanReport(gomock.Eq(2), gomock.Eq(sast.ScanReportTypeXML), gomock.Any()).
			Return([]byte("2"), nil).
			AnyTimes()
		exporter := export2.NewMockExporter(gomock.NewController(t))
		exporter.EXPECT().AddFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		result := fetchResultsData(client, exporter, 10, 3, time.Millisecond, time.Millisecond)

		assert.NoError(t, result)
	})
	t.Run("fails if triage scans returns error", func(t *testing.T) {
		projectPage := []sast.ProjectWithLastScanID{
			{ID: 1, LastScanID: 1},
			{ID: 2, LastScanID: 2},
		}
		client := sast2.NewMockClient(gomock.NewController(t))
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(0), gomock.Eq(resultsPageLimit)).
			Return(&projectPage, nil).
			AnyTimes()
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(resultsPageLimit), gomock.Eq(resultsPageLimit)).
			Return(&[]sast.ProjectWithLastScanID{}, nil).
			AnyTimes()
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(1)).
			Return(nil, fmt.Errorf("failed getting triaged scan")).
			AnyTimes()
		exporter := export2.NewMockExporter(gomock.NewController(t))
		result := fetchResultsData(client, exporter, 10, 3, time.Millisecond, time.Millisecond)

		assert.EqualError(t, result, "failed getting triaged scan")
	})
	t.Run("doesn't fail if some results fail to fetch", func(t *testing.T) {
		projectPage := []sast.ProjectWithLastScanID{
			{ID: 1, LastScanID: 1},
			{ID: 2, LastScanID: 2},
		}
		client := sast2.NewMockClient(gomock.NewController(t))
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(0), gomock.Eq(resultsPageLimit)).
			Return(&projectPage, nil).
			AnyTimes()
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(resultsPageLimit), gomock.Eq(resultsPageLimit)).
			Return(&[]sast.ProjectWithLastScanID{}, nil).
			AnyTimes()
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(1)).
			Return(&[]sast.TriagedScanResult{{ID: 1}}, nil).
			AnyTimes()
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(2)).
			Return(&[]sast.TriagedScanResult{{ID: 2}}, nil).
			AnyTimes()
		client.EXPECT().CreateScanReport(gomock.Eq(1), gomock.Eq(sast.ScanReportTypeXML), gomock.Any()).
			Return([]byte("1"), nil).
			AnyTimes()
		client.EXPECT().CreateScanReport(gomock.Eq(2), gomock.Eq(sast.ScanReportTypeXML), gomock.Any()).
			Return([]byte{}, fmt.Errorf("failed getting report #2")).
			AnyTimes()
		exporter := export2.NewMockExporter(gomock.NewController(t))
		exporter.EXPECT().AddFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		result := fetchResultsData(client, exporter, 10, 3, time.Millisecond, time.Millisecond)

		assert.NoError(t, result)
	})
}

func TestFetchSelectedData(t *testing.T) {
	t.Run("export users success case", func(t *testing.T) {
		client := sast2.NewMockClient(gomock.NewController(t))
		exporter := export2.NewMockExporter(gomock.NewController(t))
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		args := Args{
			Export:              []string{"users"},
			ProjectsActiveSince: 100,
		}

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond)

		assert.NoError(t, result)
	})
	t.Run("export users fails if fetch or write fails", func(t *testing.T) {
		client := sast2.NewMockClient(gomock.NewController(t))
		exporter := export2.NewMockExporter(gomock.NewController(t))
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.UsersFileName), gomock.Any()).
			Return(nil)
		exporter.EXPECT().AddFileWithDataSource(gomock.Eq(export.RolesFileName), gomock.Any()).
			Return(fmt.Errorf("failed fetching roles"))
		args := Args{
			Export:              []string{"users"},
			ProjectsActiveSince: 100,
		}

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond)

		assert.EqualError(t, result, "failed fetching roles")
	})
	t.Run("export users and teams success case", func(t *testing.T) {
		client := sast2.NewMockClient(gomock.NewController(t))
		exporter := export2.NewMockExporter(gomock.NewController(t))
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		args := Args{
			Export:              []string{"users", "teams"},
			ProjectsActiveSince: 100,
		}

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond)

		assert.NoError(t, result)
	})
	t.Run("export users and teams fail if fetch or write fails", func(t *testing.T) {
		client := sast2.NewMockClient(gomock.NewController(t))
		exporter := export2.NewMockExporter(gomock.NewController(t))
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

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond)

		assert.EqualError(t, result, "failed fetching LDAP team mappings")
	})
	t.Run("export users, teams and results success case", func(t *testing.T) {
		client := sast2.NewMockClient(gomock.NewController(t))
		projectPage := []sast.ProjectWithLastScanID{
			{ID: 1, LastScanID: 1},
			{ID: 2, LastScanID: 2},
		}
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(0), gomock.Any()).
			Return(&projectPage, nil)
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&[]sast.ProjectWithLastScanID{}, nil)
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(1)).
			Return(&[]sast.TriagedScanResult{{ID: 1}}, nil)
		client.EXPECT().
			GetTriagedResultsByScanID(gomock.Eq(2)).
			Return(&[]sast.TriagedScanResult{{ID: 2}}, nil)
		client.EXPECT().
			CreateScanReport(gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]byte("test"), nil).
			AnyTimes()
		exporter := export2.NewMockExporter(gomock.NewController(t))
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		exporter.EXPECT().AddFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		args := Args{
			Export:              []string{"users", "teams", "results"},
			ProjectsActiveSince: 100,
		}

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond)

		assert.NoError(t, result)
	})
	t.Run("export users, teams and results fails if result processing fails", func(t *testing.T) {
		client := sast2.NewMockClient(gomock.NewController(t))
		client.EXPECT().
			GetProjectsWithLastScanID(gomock.Any(), gomock.Eq(0), gomock.Any()).
			Return(&[]sast.ProjectWithLastScanID{}, fmt.Errorf("failed fetching projects"))
		exporter := export2.NewMockExporter(gomock.NewController(t))
		exporter.EXPECT().AddFileWithDataSource(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		args := Args{
			Export:              []string{"users", "teams", "results"},
			ProjectsActiveSince: 100,
		}

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond)

		assert.EqualError(t, result, "error searching for results")
	})
	t.Run("empty export if no export options selected", func(t *testing.T) {
		client := sast2.NewMockClient(gomock.NewController(t))
		exporter := export2.NewMockExporter(gomock.NewController(t))
		args := Args{
			Export:              []string{},
			ProjectsActiveSince: 100,
		}

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond)

		assert.NoError(t, result)
	})
	t.Run("empty export if export options are invalid", func(t *testing.T) {
		client := sast2.NewMockClient(gomock.NewController(t))
		exporter := export2.NewMockExporter(gomock.NewController(t))
		args := Args{
			Export:              []string{"test1", "test2"},
			ProjectsActiveSince: 100,
		}

		result := fetchSelectedData(client, exporter, &args, 3, time.Millisecond, time.Millisecond)

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
		exporter := export2.NewMockExporter(gomock.NewController(t))
		exporter.EXPECT().GetTmpDir().Return("/path/to/tmp/folder").MinTimes(1).MaxTimes(1)
		exporter.EXPECT().CreateExportPackage(gomock.Eq(args.ProductName), gomock.Eq(args.OutputPath)).
			Return("/path/to/output/export.zip", nil).
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
		exporter := export2.NewMockExporter(gomock.NewController(t))
		exporter.EXPECT().GetTmpDir().Return("/path/to/tmp/folder").MinTimes(1).MaxTimes(1)
		exporter.EXPECT().CreateExportPackage(gomock.Eq(args.ProductName), gomock.Eq(args.OutputPath)).
			Return("", fmt.Errorf("failed creating export package")).
			MinTimes(1).
			MaxTimes(1)

		fileName, err := exportResultsToFile(&args, exporter)

		assert.EqualError(t, err, "failed creating export package")
		assert.Equal(t, "", fileName)
	})
}
