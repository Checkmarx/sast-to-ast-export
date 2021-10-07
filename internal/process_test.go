package internal

import (
	"fmt"
	"testing"

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
		tests := []struct {
			mockExpects usersExpect
			expectedErr error
		}{
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
		tests := []struct {
			fetchMockExpects usersExpect
			writeMockExpects usersExpect
			expectedErr      error
		}{
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
		tests := []struct {
			mockExpects teamsExpect
			expectedErr error
		}{
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
		tests := []struct {
			fetchMockExpects teamsExpect
			writeMockExpects teamsExpect
			expectedErr      error
		}{
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
			client.EXPECT().
				GetTriagedResultsByScanID(gomock.Eq(k)).
				Return(&v.value, v.err). //nolint:gosec
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
