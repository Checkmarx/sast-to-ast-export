package export

import (
	"strings"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/rest"
)

func TransformTeams(teams []*rest.Team) []*rest.Team {
	var out []*rest.Team
	for _, e := range teams {
		e.ParendID = 0
		e.Name = strings.ReplaceAll(strings.TrimLeft(e.FullName, "/"), "/", "_")
		e.FullName = "/" + e.Name
		out = append(out, e)
	}
	return out
}

func TransformUsers(users []*rest.User, teams []*rest.Team) []*rest.User {
	var out []*rest.User
	for _, e := range users {
		for _, teamID := range e.TeamIDs {
			e.TeamIDs = append(e.TeamIDs, getAllChildTeamIDs(teamID, teams)...)
		}
		out = append(out, e)
	}
	return out
}

func TransformSamlTeamMappings(samlTeamMappings []*rest.SamlTeamMapping) []*rest.SamlTeamMapping {
	var out []*rest.SamlTeamMapping
	for _, e := range samlTeamMappings {
		e.TeamFullPath = "/" + strings.ReplaceAll(strings.TrimLeft(e.TeamFullPath, "/"), "/", "_")
		out = append(out, e)
	}
	return out
}

func getAllChildTeamIDs(root int, teams []*rest.Team) []int {
	var out []int
	for _, e := range teams {
		if e.ParendID == root {
			out = append(out, e.ID)
			out = append(out, getAllChildTeamIDs(e.ID, teams)...)
		}
	}
	return out
}
