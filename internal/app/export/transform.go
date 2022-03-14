package export

import (
	"strings"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/rest"
)

func TransformTeams(teams []rest.Team) []rest.Team {
	var out []rest.Team
	for _, e := range teams {
		e.ParendID = 0
		e.Name = strings.ReplaceAll(strings.TrimLeft(e.FullName, "/"), "/", "_")
		e.FullName = "/" + e.Name
		out = append(out, e)
	}
	return out
}
