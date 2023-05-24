package export

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/common"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/rest"
)

const (
	InstallationEngineServiceName = "Checkmarx Engine Service"
	installationScansManagerName  = "Checkmarx Scans Manager"
	installationContentPackName   = "Checkmarx Queries Pack"
	engineServersStatusOffline    = "Offline"
)

type TransformOptions struct {
	NestedTeams bool
}

// TransformTeams flattens teams.
func TransformTeams(teams []*rest.Team, options TransformOptions) []*rest.Team {
	if options.NestedTeams {
		return teams
	}
	out := make([]*rest.Team, 0)
	for _, e := range teams {
		e.ParendID = 0
		e.Name = strings.ReplaceAll(strings.TrimLeft(e.FullName, "/"), "/", "_")
		e.FullName = "/" + e.Name
		out = append(out, e)
	}
	return out
}

// TransformUsers reassigns users in the context of flatten teams.
// Note "teams" list passed must be the original, non-flattened, list
func TransformUsers(users []*rest.User, teams []*rest.Team, options TransformOptions) []*rest.User {
	if options.NestedTeams {
		return users
	}
	out := make([]*rest.User, 0)
	for _, e := range users {
		for _, teamID := range e.TeamIDs {
			e.TeamIDs = append(e.TeamIDs, getAllChildTeamIDs(teamID, teams)...)
		}
		out = append(out, e)
	}
	return out
}

// TransformSamlTeamMappings updates team mapping in the context of flatten teams.
func TransformSamlTeamMappings(samlTeamMappings []*rest.SamlTeamMapping, options TransformOptions) []*rest.SamlTeamMapping {
	if options.NestedTeams {
		return samlTeamMappings
	}
	out := make([]*rest.SamlTeamMapping, 0)
	for _, e := range samlTeamMappings {
		e.TeamFullPath = "/" + strings.ReplaceAll(strings.TrimLeft(e.TeamFullPath, "/"), "/", "_")
		out = append(out, e)
	}
	return out
}

// TransformXMLInstallationMappings updates installation mapping.
func TransformXMLInstallationMappings(installationMappings *soap.GetInstallationSettingsResponse) []*common.InstallationMapping {
	out := make([]*common.InstallationMapping, 0)
	if installationMappings == nil {
		return []*common.InstallationMapping{}
	}

	for _, e := range installationMappings.GetInstallationSettingsResult.InstallationSettingsList.InstallationSetting {
		if e.Name == InstallationEngineServiceName || e.Name == installationScansManagerName {
			if !ContainsEngine(InstallationEngineServiceName, out) {
				out = append(out, &common.InstallationMapping{
					Name:    InstallationEngineServiceName,
					Version: e.Version,
					Hotfix:  e.Hotfix,
				})
			}
		} else if e.Name == installationContentPackName {
			out = append(out, &common.InstallationMapping{
				Name:    e.Name,
				Version: e.Version,
				Hotfix:  e.Hotfix,
			})
		}
	}
	return out
}

// TransformScanReport updates scan report in context of flatten teams.
func TransformScanReport(xml []byte, options TransformOptions) ([]byte, error) {
	if options.NestedTeams {
		return xml, nil
	}
	var teamPath string
	out := replaceKeyValue(xml, "TeamFullPathOnReportDate", func(s string) string {
		teamPath = strings.ReplaceAll(s, "\\", "_")
		return teamPath
	})
	out = replaceKeyValue(out, "Team", func(s string) string {
		return teamPath
	})
	return out, nil
}

// TransformEngineServers just for SAST distributed architecture just for 9.4 or higher.
func TransformEngineServers(servers []*rest.EngineServer) []*common.InstallationMapping {
	out := make([]*common.InstallationMapping, 0)
	if servers == nil {
		return []*common.InstallationMapping{}
	}

	if len(servers) == 1 {
		out = append(out, &common.InstallationMapping{
			Name:    InstallationEngineServiceName,
			Version: servers[0].CxVersion,
			Hotfix:  "",
		})
	} else if len(servers) > 1 {
		for _, e := range servers {
			if e.Status.Value != engineServersStatusOffline {
				if !ContainsEngine(InstallationEngineServiceName, out) {
					out = append(out, &common.InstallationMapping{
						Name:    InstallationEngineServiceName,
						Version: e.CxVersion,
						Hotfix:  "",
					})
				}
			}
		}
	}

	return out
}

// ContainsEngine returns true if already exists object filtered by name.
func ContainsEngine(needle string, data []*common.InstallationMapping) bool {
	for _, v := range data {
		if needle == v.Name {
			return true
		}
	}
	return false
}

// getAllChildTeamIDs returns all child team ids relative to a root team id.
func getAllChildTeamIDs(root int, teams []*rest.Team) []int {
	out := make([]int, 0)
	for _, e := range teams {
		if e.ParendID == root {
			out = append(out, e.ID)
			out = append(out, getAllChildTeamIDs(e.ID, teams)...)
		}
	}
	return out
}

// replaceKeyValue
func replaceKeyValue(d []byte, key string, getValue func(string) string) []byte {
	re := regexp.MustCompile(fmt.Sprintf(`(%s)="([^"]+)"`, key))
	submatchCount := 2
	return re.ReplaceAllFunc(d, func(entry []byte) []byte {
		matches := re.FindAllSubmatch(entry, submatchCount)
		value := getValue(string(matches[0][2]))
		return []byte(fmt.Sprintf(`%s=%q`, key, value))
	})
}
