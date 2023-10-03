package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiClientUtils(t *testing.T) {
	type TestObj struct {
		fromDate, teamName, projectIds, expectedResult string
	}
	fromDate := "2022-01-15"

	t.Run("Test filter for project list", func(t *testing.T) {
		tests := []TestObj{
			{
				fromDate:       fromDate,
				teamName:       "",
				projectIds:     "",
				expectedResult: "CreatedDate gt 2022-01-15",
			},
			{
				fromDate:       fromDate,
				teamName:       "TestTeam",
				projectIds:     "",
				expectedResult: "CreatedDate gt 2022-01-15 and OwningTeam/FullName eq 'TestTeam'",
			},
			{
				fromDate:       fromDate,
				teamName:       "",
				projectIds:     "1",
				expectedResult: "CreatedDate gt 2022-01-15 and Id eq 1",
			},
			{
				fromDate:       fromDate,
				teamName:       "TestTeam",
				projectIds:     "1,2",
				expectedResult: "CreatedDate gt 2022-01-15 and OwningTeam/FullName eq 'TestTeam' and Id in (1,2)",
			},
			{
				fromDate:       fromDate,
				teamName:       "",
				projectIds:     "1-5",
				expectedResult: "CreatedDate gt 2022-01-15 and Id ge 1 and Id le 5",
			},
			{
				fromDate:       fromDate,
				teamName:       "",
				projectIds:     "wrong_num",
				expectedResult: "CreatedDate gt 2022-01-15 and Id gt 0",
			},
			{
				fromDate:       "",
				teamName:       "",
				projectIds:     "1,2",
				expectedResult: "Id in (1,2)",
			},
			{
				fromDate:       "",
				teamName:       "TestTeam",
				projectIds:     "1,2",
				expectedResult: "Id in (1,2) and OwningTeam/FullName eq 'TestTeam'",
			},
		}

		for _, test := range tests {
			result := GetFilterForProjects(test.fromDate, test.teamName, test.projectIds)
			assert.Equal(t, test.expectedResult, result)
		}
	})

	t.Run("Test filter for project list with last scan", func(t *testing.T) {
		tests := []TestObj{
			{
				fromDate:       fromDate,
				teamName:       "",
				projectIds:     "",
				expectedResult: "LastScan/ScanCompletedOn gt 2022-01-15",
			},
			{
				fromDate:       fromDate,
				teamName:       "TestTeam",
				projectIds:     "",
				expectedResult: "LastScan/ScanCompletedOn gt 2022-01-15 and OwningTeam/FullName eq 'TestTeam'",
			},
			{
				fromDate:       fromDate,
				teamName:       "",
				projectIds:     "5-2",
				expectedResult: "LastScan/ScanCompletedOn gt 2022-01-15 and Id ge 2 and Id le 5",
			},
			{
				fromDate:       fromDate,
				teamName:       "TestTeam",
				projectIds:     "1,2",
				expectedResult: "LastScan/ScanCompletedOn gt 2022-01-15 and OwningTeam/FullName eq 'TestTeam' and Id in (1,2)",
			},
			{
				fromDate:       "",
				teamName:       "",
				projectIds:     "1,2",
				expectedResult: "Id in (1,2)",
			},
			{
				fromDate:       "",
				teamName:       "TestTeam",
				projectIds:     "1,2",
				expectedResult: "Id in (1,2) and OwningTeam/FullName eq 'TestTeam'",
			},
		}

		for _, test := range tests {
			result := GetFilterForProjectsWithLastScan(test.fromDate, test.teamName, test.projectIds)
			assert.Equal(t, test.expectedResult, result)
		}
	})
}
