package export

import (
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/rest"
	"github.com/stretchr/testify/assert"
)

type transformTeamTest struct {
	Name     string
	Input    []rest.Team
	Expected []rest.Team
}

func TestTransformTeams(t *testing.T) {
	tests := []transformTeamTest{
		{"empty input", []rest.Team{}, []rest.Team{}},
		{
			"one root team",
			[]rest.Team{{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0}},
			[]rest.Team{{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0}},
		},
		{
			"one sub-level",
			[]rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
				{ID: 3, Name: "TeamC", FullName: "/TeamA/TeamC", ParendID: 1},
			},
			[]rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamA_TeamB", FullName: "/TeamA_TeamB", ParendID: 0},
				{ID: 3, Name: "TeamA_TeamC", FullName: "/TeamA_TeamC", ParendID: 0},
			},
		},
		{
			"two sub-levels",
			[]rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
				{ID: 3, Name: "TeamC", FullName: "/TeamA/TeamB/TeamC", ParendID: 2},
			},
			[]rest.Team{
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamA_TeamB", FullName: "/TeamA_TeamB", ParendID: 0},
				{ID: 3, Name: "TeamA_TeamB_TeamC", FullName: "/TeamA_TeamB_TeamC", ParendID: 0},
			},
		},
		{
			"two trees with 3 sub-levels",
			[]rest.Team{ //nolint:dupl
				{ID: 1, Name: "TeamA", FullName: "/TeamA", ParendID: 0},
				{ID: 2, Name: "TeamB", FullName: "/TeamA/TeamB", ParendID: 1},
				{ID: 3, Name: "TeamC", FullName: "/TeamA/TeamB/TeamC", ParendID: 2},
				{ID: 4, Name: "TeamD", FullName: "/TeamA/TeamB/TeamC/TeamD", ParendID: 3},
				{ID: 5, Name: "TeamE", FullName: "/TeamE", ParendID: 0},
				{ID: 6, Name: "TeamF", FullName: "/TeamE/TeamF", ParendID: 5},
				{ID: 7, Name: "TeamG", FullName: "/TeamE/TeamF/TeamG", ParendID: 5},
				{ID: 8, Name: "TeamH", FullName: "/TeamE/TeamF/TeamG/TeamH", ParendID: 5},
			},
			[]rest.Team{ //nolint:dupl
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
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := TransformTeams(test.Input)
			assert.ElementsMatch(t, test.Expected, result)
		})
	}
}
