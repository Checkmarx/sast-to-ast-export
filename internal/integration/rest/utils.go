package rest

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Retry struct {
	Attempts int
	MinSleep,
	MaxSleep time.Duration
}

// GetFilterForProjects get filter string for projects list
func GetFilterForProjects(fromDate, teamName, projectIDs string) string {
	if teamName == "" && projectIDs == "" {
		return fmt.Sprintf("CreatedDate gt %s", fromDate)
	}
	if fromDate == "" {
		return getProjectFilterForEmptyDate(projectIDs, teamName)
	}
	if teamName == "" {
		return fmt.Sprintf("CreatedDate gt %s and %s", fromDate, getProjectIDsFilter(projectIDs))
	}
	if projectIDs == "" {
		return fmt.Sprintf("CreatedDate gt %s and %s", fromDate, getTeamFilter(teamName))
	}
	return fmt.Sprintf("CreatedDate gt %s and %s and %s", fromDate, getTeamFilter(teamName),
		getProjectIDsFilter(projectIDs))
}

// GetFilterForProjectsWithLastScan get filter string for projects list with last scan
func GetFilterForProjectsWithLastScan(fromDate, teamName, projectIDs string) string {
	if teamName == "" && projectIDs == "" {
		return fmt.Sprintf("LastScan/ScanCompletedOn gt %s", fromDate)
	}
	if fromDate == "" {
		return getProjectFilterForEmptyDate(projectIDs, teamName)
	}
	if teamName == "" {
		return fmt.Sprintf("LastScan/ScanCompletedOn gt %s and %s", fromDate, getProjectIDsFilter(projectIDs))
	}
	if projectIDs == "" {
		return fmt.Sprintf("LastScan/ScanCompletedOn gt %s and %s", fromDate, getTeamFilter(teamName))
	}
	return fmt.Sprintf("LastScan/ScanCompletedOn gt %s and %s and %s", fromDate, getTeamFilter(teamName),
		getProjectIDsFilter(projectIDs))
}

// getProjectFilterForEmptyDate get project filter when date empty
func getProjectFilterForEmptyDate(projectIDs, teamName string) string {
	if teamName == "" {
		return getProjectIDsFilter(projectIDs)
	}
	return fmt.Sprintf("%s and %s", getProjectIDsFilter(projectIDs), getTeamFilter(teamName))
}

// getTeamFilter get filter string for team
func getTeamFilter(teamName string) string {
	return fmt.Sprintf("OwningTeam/FullName eq '%s'", teamName)
}

// getProjectIDsFilter get filter string for project-id option
func getProjectIDsFilter(projectIDs string) string {
	if matched, _ := regexp.MatchString(`^\d+$`, projectIDs); matched {
		return fmt.Sprintf("Id eq %s", projectIDs)
	}
	if matched, _ := regexp.MatchString(`^\d+(,\s?\d+)+$`, projectIDs); matched {
		return fmt.Sprintf("Id in (%s)", projectIDs)
	}
	if matched, _ := regexp.MatchString(`^\d+\s?-\s?\d+$`, projectIDs); matched {
		ids := strings.Split(projectIDs, "-")
		minValue, maxValue := getMinMax(ids)
		return fmt.Sprintf("Id ge %d and Id le %d", minValue, maxValue)
	}

	log.Warn().Msg("--project-id has wrong param. It should be like --project-id 1 or 1,3,8 or 1-3")
	return "Id gt 0"
}

// getMinMax get min and max values
func getMinMax(ids []string) (minValue, maxValue int) {
	minValue, _ = strconv.Atoi(strings.Trim(ids[0], " "))
	maxValue, _ = strconv.Atoi(strings.Trim(ids[1], " "))
	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}

	return minValue, maxValue
}
