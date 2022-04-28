package export

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/rest"
)

// TransformTeams flattens teams.
func TransformTeams(teams []*rest.Team) []*rest.Team {
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
func TransformUsers(users []*rest.User, teams []*rest.Team) []*rest.User {
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
func TransformSamlTeamMappings(samlTeamMappings []*rest.SamlTeamMapping) []*rest.SamlTeamMapping {
	out := make([]*rest.SamlTeamMapping, 0)
	for _, e := range samlTeamMappings {
		e.TeamFullPath = "/" + strings.ReplaceAll(strings.TrimLeft(e.TeamFullPath, "/"), "/", "_")
		out = append(out, e)
	}
	return out
}

// TransformScanReport updates scan report in context of flatten teams.
func TransformScanReport(xml []byte) ([]byte, error) {
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
