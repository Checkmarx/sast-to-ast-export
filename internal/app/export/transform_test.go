package export

import (
	"fmt"
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/common"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/rest"
	"github.com/stretchr/testify/assert"
)

type transformTeamsTest struct {
	Name     string
	Input    []*rest.Team
	Options  TransformOptions
	Expected []*rest.Team
}

type transformUsersTest struct {
	Name     string
	Input    []*rest.User
	Teams    []*rest.Team
	Options  TransformOptions
	Expected []*rest.User
}

func TestTransformTeams(t *testing.T) {
	tests := []transformTeamsTest{
		{"empty input", []*rest.Team{}, TransformOptions{}, []*rest.Team{}},
		{
			"one root team",
			[]*rest.Team{{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0}},
			TransformOptions{},
			[]*rest.Team{{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0}},
		},
		{
			"one sub-level",
			[]*rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
				{ID: 3, Name: "TeamC", FullName: "/TeamA/TeamC", ParendID: 1},
			},
			TransformOptions{},
			[]*rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamA_TeamB", FullName: "/TeamA_TeamB", ParendID: 0},
				{ID: 3, Name: "TeamA_TeamC", FullName: "/TeamA_TeamC", ParendID: 0},
			},
		},
		{
			"two sub-levels",
			[]*rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
				{ID: 3, Name: "TeamC", FullName: "/TeamA/TeamB/TeamC", ParendID: 2},
			},
			TransformOptions{},
			[]*rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamA_TeamB", FullName: "/TeamA_TeamB", ParendID: 0},
				{ID: 3, Name: "TeamA_TeamB_TeamC", FullName: "/TeamA_TeamB_TeamC", ParendID: 0},
			},
		},
		{
			"two trees with 3 sub-levels",
			[]*rest.Team{ //nolint:dupl
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
				{ID: 3, Name: "TeamC", FullName: "/TeamA/TeamB/TeamC", ParendID: 2},
				{ID: 4, Name: "TeamD", FullName: "/TeamA/TeamB/TeamC/TeamD", ParendID: 3},
				{ID: 5, Name: "TeamE", FullName: "/TeamE", ParendID: 0},
				{ID: 6, Name: "TeamF", FullName: "/TeamE/TeamF", ParendID: 5},
				{ID: 7, Name: "TeamG", FullName: "/TeamE/TeamF/TeamG", ParendID: 5},
				{ID: 8, Name: "TeamH", FullName: "/TeamE/TeamF/TeamG/TeamH", ParendID: 5},
			},
			TransformOptions{},
			[]*rest.Team{ //nolint:dupl
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamA_TeamB", FullName: "/TeamA_TeamB", ParendID: 0},
				{ID: 3, Name: "TeamA_TeamB_TeamC", FullName: "/TeamA_TeamB_TeamC", ParendID: 0},
				{ID: 4, Name: "TeamA_TeamB_TeamC_TeamD", FullName: "/TeamA_TeamB_TeamC_TeamD", ParendID: 0},
				{ID: 5, Name: "TeamE", FullName: "/TeamE", ParendID: 0},
				{ID: 6, Name: "TeamE_TeamF", FullName: "/TeamE_TeamF", ParendID: 0},
				{ID: 7, Name: "TeamE_TeamF_TeamG", FullName: "/TeamE_TeamF_TeamG", ParendID: 0},
				{ID: 8, Name: "TeamE_TeamF_TeamG_TeamH", FullName: "/TeamE_TeamF_TeamG_TeamH", ParendID: 0},
			},
		},
		{
			"nested teams enabled",
			[]*rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
				{ID: 3, Name: "TeamC", FullName: "/TeamA/TeamC", ParendID: 1},
			},
			TransformOptions{NestedTeams: true},
			[]*rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
				{ID: 3, Name: "TeamC", FullName: "/TeamA/TeamC", ParendID: 1},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := TransformTeams(test.Input, test.Options)
			assert.ElementsMatch(t, test.Expected, result)
		})
	}
}

func TestTransformUsers(t *testing.T) {
	tests := []transformUsersTest{
		{"empty input", []*rest.User{}, []*rest.Team{}, TransformOptions{}, []*rest.User{}},
		{
			"one user in root team",
			[]*rest.User{{ID: 1, UserName: "Alice", TeamIDs: []int{1}}},
			[]*rest.Team{{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0}},
			TransformOptions{},
			[]*rest.User{{ID: 1, UserName: "Alice", TeamIDs: []int{1}}},
		},
		{
			"two users in two team levels",
			[]*rest.User{
				{ID: 1, UserName: "Alice", TeamIDs: []int{1}},
				{ID: 2, UserName: "Bob", TeamIDs: []int{2}},
			},
			[]*rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
			},
			TransformOptions{},
			[]*rest.User{
				{ID: 1, UserName: "Alice", TeamIDs: []int{1, 2}},
				{ID: 2, UserName: "Bob", TeamIDs: []int{2}},
			},
		},
		{
			"three users in three team levels",
			[]*rest.User{
				{ID: 1, UserName: "Alice", TeamIDs: []int{1}},
				{ID: 2, UserName: "Bob", TeamIDs: []int{2}},
				{ID: 3, UserName: "Charlie", TeamIDs: []int{3}},
			},
			[]*rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
				{ID: 3, Name: "TeamC", FullName: "/TeamA/TeamB/TeamC", ParendID: 2},
			},
			TransformOptions{},
			[]*rest.User{
				{ID: 1, UserName: "Alice", TeamIDs: []int{1, 2, 3}},
				{ID: 2, UserName: "Bob", TeamIDs: []int{2, 3}},
				{ID: 3, UserName: "Charlie", TeamIDs: []int{3}},
			},
		},
		{
			"siz users in two team trees with multiple levels",
			[]*rest.User{
				{ID: 1, UserName: "Alice", TeamIDs: []int{1}},
				{ID: 2, UserName: "Bob", TeamIDs: []int{2}},
				{ID: 3, UserName: "Charlie", TeamIDs: []int{3}},
				{ID: 4, UserName: "Diane", TeamIDs: []int{4}},
				{ID: 5, UserName: "Emily", TeamIDs: []int{5}},
				{ID: 6, UserName: "Fred", TeamIDs: []int{6}},
			},
			[]*rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
				{ID: 3, Name: "TeamC", FullName: "/TeamA/TeamB/TeamC", ParendID: 2},
				{ID: 4, Name: "TeamD", FullName: "/TeamD", ParendID: 0},
				{ID: 5, Name: "TeamE", FullName: "/TeamD/TeamE", ParendID: 4},
				{ID: 6, Name: "TeamF", FullName: "/TeamD/TeamF", ParendID: 4},
			},
			TransformOptions{},
			[]*rest.User{
				{ID: 1, UserName: "Alice", TeamIDs: []int{1, 2, 3}},
				{ID: 2, UserName: "Bob", TeamIDs: []int{2, 3}},
				{ID: 3, UserName: "Charlie", TeamIDs: []int{3}},
				{ID: 4, UserName: "Diane", TeamIDs: []int{4, 5, 6}},
				{ID: 5, UserName: "Emily", TeamIDs: []int{5}},
				{ID: 6, UserName: "Fred", TeamIDs: []int{6}},
			},
		},
		{
			"nested teams enabled",
			[]*rest.User{
				{ID: 1, UserName: "Alice", TeamIDs: []int{1}},
				{ID: 2, UserName: "Bob", TeamIDs: []int{2}},
			},
			[]*rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
			},
			TransformOptions{NestedTeams: true},
			[]*rest.User{
				{ID: 1, UserName: "Alice", TeamIDs: []int{1}},
				{ID: 2, UserName: "Bob", TeamIDs: []int{2}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := TransformUsers(test.Input, test.Teams, test.Options)
			assert.ElementsMatch(t, test.Expected, result)
		})
	}
}

func TestTransformSamlTeamMappings(t *testing.T) {
	t.Run("no mappings", func(t *testing.T) {
		var samlTeamMappings []*rest.SamlTeamMapping

		result := TransformSamlTeamMappings(samlTeamMappings, TransformOptions{})

		var expected []*rest.SamlTeamMapping
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("top level team", func(t *testing.T) {
		samlTeamMappings := []*rest.SamlTeamMapping{
			{ID: 1, SamlIdentityProviderID: 1, TeamID: 1, TeamFullPath: "/TeamA", SamlAttributeValue: "team"},
		}

		result := TransformSamlTeamMappings(samlTeamMappings, TransformOptions{})

		expected := []*rest.SamlTeamMapping{
			{ID: 1, SamlIdentityProviderID: 1, TeamID: 1, TeamFullPath: "/TeamA", SamlAttributeValue: "team"},
		}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("team in sub-level", func(t *testing.T) {
		samlTeamMappings := []*rest.SamlTeamMapping{
			{ID: 1, SamlIdentityProviderID: 1, TeamID: 2, TeamFullPath: "/TeamA/TeamB", SamlAttributeValue: "team"},
		}

		result := TransformSamlTeamMappings(samlTeamMappings, TransformOptions{})

		expected := []*rest.SamlTeamMapping{
			{ID: 1, SamlIdentityProviderID: 1, TeamID: 2, TeamFullPath: "/TeamA_TeamB", SamlAttributeValue: "team"},
		}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("team in 2nd sub-level", func(t *testing.T) {
		samlTeamMappings := []*rest.SamlTeamMapping{
			{ID: 1, SamlIdentityProviderID: 1, TeamID: 3, TeamFullPath: "/TeamA/TeamB/TeamC", SamlAttributeValue: "team"},
		}

		result := TransformSamlTeamMappings(samlTeamMappings, TransformOptions{})

		expected := []*rest.SamlTeamMapping{
			{ID: 1, SamlIdentityProviderID: 1, TeamID: 3, TeamFullPath: "/TeamA_TeamB_TeamC", SamlAttributeValue: "team"},
		}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("nested teams enabled", func(t *testing.T) {
		samlTeamMappings := []*rest.SamlTeamMapping{
			{ID: 1, SamlIdentityProviderID: 1, TeamID: 2, TeamFullPath: "/TeamA/TeamB", SamlAttributeValue: "team"},
		}

		result := TransformSamlTeamMappings(samlTeamMappings, TransformOptions{NestedTeams: true})

		expected := []*rest.SamlTeamMapping{
			{ID: 1, SamlIdentityProviderID: 1, TeamID: 2, TeamFullPath: "/TeamA/TeamB", SamlAttributeValue: "team"},
		}
		assert.ElementsMatch(t, expected, result)
	})
}

func TestTransformScanReport(t *testing.T) {
	t.Run("root team", func(t *testing.T) {
		report := newMockScanReportXML("TeamA", "TeamA")

		result, err := TransformScanReport([]byte(report), TransformOptions{})

		assert.NoError(t, err)
		assert.Equal(t, report, string(result))
	})

	t.Run("one level deep team", func(t *testing.T) {
		report := newMockScanReportXML("TeamB", "TeamA\\TeamB")

		result, err := TransformScanReport([]byte(report), TransformOptions{})

		assert.NoError(t, err)
		expected := newMockScanReportXML("TeamA_TeamB", "TeamA_TeamB")
		assert.Equal(t, expected, string(result))
	})

	t.Run("two levels deep team", func(t *testing.T) {
		report := newMockScanReportXML("TeamC", "TeamA\\TeamB\\TeamC")

		result, err := TransformScanReport([]byte(report), TransformOptions{})

		assert.NoError(t, err)
		expected := newMockScanReportXML("TeamA_TeamB_TeamC", "TeamA_TeamB_TeamC")
		assert.Equal(t, expected, string(result))
	})

	t.Run("nested teams enabled", func(t *testing.T) {
		report := newMockScanReportXML("TeamB", "TeamA\\TeamB")

		result, err := TransformScanReport([]byte(report), TransformOptions{NestedTeams: true})

		assert.NoError(t, err)
		expected := newMockScanReportXML("TeamB", "TeamA\\TeamB")
		assert.Equal(t, expected, string(result))
	})
}

func TestTransformEngineServers(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		engineServers := []*rest.EngineServer{
			{
				ID:        1,
				Name:      "blabla1",
				URI:       "http://localhost",
				CxVersion: "9.3.4.1111",
				Status: rest.StatusEngineServer{
					ID:    1,
					Value: "Idle",
				},
			},
			{
				ID:        2,
				Name:      "blabla2",
				URI:       "http://localhost",
				CxVersion: "9.3.4.1111",
				Status: rest.StatusEngineServer{
					ID:    1,
					Value: "Offline",
				},
			},
		}

		expected := []*common.InstallationMapping{
			{
				Name:    "Checkmarx Engine Service",
				Version: "9.3.4.1111",
				Hotfix:  "",
			},
		}

		result := TransformEngineServers(engineServers)

		assert.Equal(t, expected, result)
	})

	t.Run("success without status case", func(t *testing.T) {
		engineServers := []*rest.EngineServer{
			{
				ID:        1,
				Name:      "blabla",
				URI:       "http://localhost",
				CxVersion: "9.3.4.1111",
			},
		}
		expected := []*common.InstallationMapping{
			{
				Name:    "Checkmarx Engine Service",
				Version: "9.3.4.1111",
				Hotfix:  "",
			},
		}

		result := TransformEngineServers(engineServers)

		assert.Equal(t, expected, result)
	})

	t.Run("success empty case", func(t *testing.T) {
		expected := []*common.InstallationMapping{}

		result := TransformEngineServers(nil)

		assert.Equal(t, expected, result)
	})
}

func TestGetAllChildTeamIDs(t *testing.T) {
	teams := []*rest.Team{
		{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
		{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
		{ID: 3, Name: "TeamC", FullName: "/TeamC", ParendID: 0},
		{ID: 4, Name: "TeamD", FullName: "/TeamC/TeamD", ParendID: 3},
		{ID: 5, Name: "TeamE", FullName: "/TeamC/TeamD/TeamE", ParendID: 4},
		{ID: 6, Name: "TeamF", FullName: "/TeamC/TeamF", ParendID: 3},
	}

	t.Run("starting from TeamA", func(t *testing.T) {
		result := getAllChildTeamIDs(1, teams)

		expected := []int{2}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("starting from TeamC", func(t *testing.T) {
		result := getAllChildTeamIDs(3, teams)

		expected := []int{4, 5, 6}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("starting from TeamD", func(t *testing.T) {
		result := getAllChildTeamIDs(4, teams)

		expected := []int{5}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("starting from TeamF", func(t *testing.T) {
		result := getAllChildTeamIDs(6, teams)

		var expected []int
		assert.ElementsMatch(t, expected, result)
	})
}

func TestReplaceKeyValue(t *testing.T) {
	s := `a="1" b="2" c="3"`

	t.Run("a", func(t *testing.T) {
		result := replaceKeyValue([]byte(s), "a", func(a string) string {
			return fmt.Sprintf(".%s.", a)
		})

		expected := `a=".1." b="2" c="3"`
		assert.Equal(t, expected, string(result))
	})

	t.Run("b", func(t *testing.T) {
		result := replaceKeyValue([]byte(s), "b", func(b string) string {
			return fmt.Sprintf("-%s-", b)
		})

		expected := `a="1" b="-2-" c="3"`
		assert.Equal(t, expected, string(result))
	})
}

// nolint
func TestTransformXMLInstallationMappings(t *testing.T) {
	engineService := &soap.InstallationSetting{
		Name:    "Checkmarx Engine Service",
		Version: "9.3.4.1111",
		Hotfix:  "Hotfix",
	}
	scansManager := &soap.InstallationSetting{
		Name:    "Checkmarx Scans Manager",
		Version: "9.3.0.1111",
		Hotfix:  "Hotfix",
	}
	queriesPack := &soap.InstallationSetting{
		Name:    "Checkmarx Queries Pack",
		Version: "9.3.4.5111",
		Hotfix:  "Hotfix",
	}
	t.Run("no installations", func(t *testing.T) {
		var installationMappings *soap.GetInstallationSettingsResponse

		result := TransformXMLInstallationMappings(installationMappings)

		var expected []*common.InstallationMapping
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("only engine service", func(t *testing.T) {
		soapResponseSuccess := soap.GetInstallationSettingsResponse{
			GetInstallationSettingsResult: soap.GetInstallationSettingsResult{
				IsSuccesfull: "true",
				InstallationSettingsList: soap.InstallationSettingsList{
					InstallationSetting: []*soap.InstallationSetting{
						engineService,
					},
				},
			},
		}

		result := TransformXMLInstallationMappings(&soapResponseSuccess)
		expected := []*common.InstallationMapping{
			{
				Name:    "Checkmarx Engine Service",
				Version: "9.3.4.1111",
				Hotfix:  "Hotfix",
			},
		}

		assert.ElementsMatch(t, expected, result)
	})

	t.Run("only queries pack", func(t *testing.T) {
		soapResponseSuccess := soap.GetInstallationSettingsResponse{
			GetInstallationSettingsResult: soap.GetInstallationSettingsResult{
				IsSuccesfull: "true",
				InstallationSettingsList: soap.InstallationSettingsList{
					InstallationSetting: []*soap.InstallationSetting{
						queriesPack,
					},
				},
			},
		}

		result := TransformXMLInstallationMappings(&soapResponseSuccess)
		expected := []*common.InstallationMapping{
			{
				Name:    "Checkmarx Queries Pack",
				Version: "9.3.4.5111",
				Hotfix:  "Hotfix",
			},
		}

		assert.ElementsMatch(t, expected, result)
	})

	t.Run("both engine service and queries pack", func(t *testing.T) {
		soapResponseSuccess := soap.GetInstallationSettingsResponse{
			GetInstallationSettingsResult: soap.GetInstallationSettingsResult{
				IsSuccesfull: "true",
				InstallationSettingsList: soap.InstallationSettingsList{
					InstallationSetting: []*soap.InstallationSetting{
						engineService,
						queriesPack,
					},
				},
			},
		}

		result := TransformXMLInstallationMappings(&soapResponseSuccess)
		expected := []*common.InstallationMapping{
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
		}

		assert.ElementsMatch(t, expected, result)
	})

	t.Run("all options engine service, scans manager and queries pack", func(t *testing.T) {
		soapResponseSuccess := soap.GetInstallationSettingsResponse{
			GetInstallationSettingsResult: soap.GetInstallationSettingsResult{
				IsSuccesfull: "true",
				InstallationSettingsList: soap.InstallationSettingsList{
					InstallationSetting: []*soap.InstallationSetting{
						scansManager,
						engineService,
						queriesPack,
					},
				},
			},
		}

		result := TransformXMLInstallationMappings(&soapResponseSuccess)
		expected := []*common.InstallationMapping{
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
		}

		assert.ElementsMatch(t, expected, result)
	})

	t.Run("all options engine service, scans manager and queries pack inverted order", func(t *testing.T) {
		soapResponseSuccess := soap.GetInstallationSettingsResponse{
			GetInstallationSettingsResult: soap.GetInstallationSettingsResult{
				IsSuccesfull: "true",
				InstallationSettingsList: soap.InstallationSettingsList{
					InstallationSetting: []*soap.InstallationSetting{
						engineService,
						scansManager,
						queriesPack,
					},
				},
			},
		}

		result := TransformXMLInstallationMappings(&soapResponseSuccess)
		expected := []*common.InstallationMapping{
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
		}

		assert.ElementsMatch(t, expected, result)
	})

	t.Run("only both engine service and manager first", func(t *testing.T) {
		soapResponseSuccess := soap.GetInstallationSettingsResponse{
			GetInstallationSettingsResult: soap.GetInstallationSettingsResult{
				IsSuccesfull: "true",
				InstallationSettingsList: soap.InstallationSettingsList{
					InstallationSetting: []*soap.InstallationSetting{
						scansManager,
						engineService,
					},
				},
			},
		}

		result := TransformXMLInstallationMappings(&soapResponseSuccess)
		expected := []*common.InstallationMapping{
			{
				Name:    "Checkmarx Engine Service",
				Version: "9.3.4.1111",
				Hotfix:  "Hotfix",
			},
		}

		assert.ElementsMatch(t, expected, result)
	})

	t.Run("only both engine service and manager pack inverted order", func(t *testing.T) {
		soapResponseSuccess := soap.GetInstallationSettingsResponse{
			GetInstallationSettingsResult: soap.GetInstallationSettingsResult{
				IsSuccesfull: "true",
				InstallationSettingsList: soap.InstallationSettingsList{
					InstallationSetting: []*soap.InstallationSetting{
						engineService,
						scansManager,
					},
				},
			},
		}

		result := TransformXMLInstallationMappings(&soapResponseSuccess)
		expected := []*common.InstallationMapping{
			{
				Name:    "Checkmarx Engine Service",
				Version: "9.3.4.1111",
				Hotfix:  "Hotfix",
			},
		}

		assert.ElementsMatch(t, expected, result)
	})
}

func newMockScanReportXML(teamName, teamFullPath string) string {
	// nolint:lll
	return fmt.Sprintf(`
<?xml version="1.0" encoding="utf-8"?>
<CxXMLResults InitiatorName="test" Owner="test" ScanId="1000000" ProjectId="1" ProjectName="test" TeamFullPathOnReportDate="%s" DeepLink="http://localhost/CxWebClient/ViewerMain.aspx?scanid=1000000&amp;projectid=1" ScanStart="Thursday, September 30, 2021 11:57:20 AM" Preset="Checkmarx Default" ScanTime="00h:00m:44s" LinesOfCodeScanned="1330" FilesScanned="9" ReportCreationTime="Thursday, February 24, 2022 4:15:03 PM" Team="%s" CheckmarxVersion="9.3.0.1139" ScanComments="" ScanType="Full" SourceOrigin="LocalPath" Visibility="Public">

</CxXMLResults>
`, teamFullPath, teamName)
}
