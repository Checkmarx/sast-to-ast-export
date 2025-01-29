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
		return fmt.Sprintf("CreatedDate gt %s and %s", fromDate, getProjectIdsFilter(projectIDs))
	}
	if projectIDs == "" {
		return fmt.Sprintf("CreatedDate gt %s and %s", fromDate, getTeamFilter(teamName))
	}
	return fmt.Sprintf("CreatedDate gt %s and %s and %s", fromDate, getTeamFilter(teamName),
		getProjectIdsFilter(projectIDs))
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
		return fmt.Sprintf("LastScan/ScanCompletedOn gt %s and %s", fromDate, getProjectIdsFilter(projectIDs))
	}
	if projectIDs == "" {
		return fmt.Sprintf("LastScan/ScanCompletedOn gt %s and %s", fromDate, getTeamFilter(teamName))
	}
	return fmt.Sprintf("LastScan/ScanCompletedOn gt %s and %s and %s", fromDate, getTeamFilter(teamName),
		getProjectIdsFilter(projectIDs))
}

// getProjectFilterForEmptyDate get project filter when date empty
func getProjectFilterForEmptyDate(projectIDs, teamName string) string {
	if teamName == "" {
		return getProjectIdsFilter(projectIDs)
	}
	return fmt.Sprintf("%s and %s", getProjectIdsFilter(projectIDs), getTeamFilter(teamName))
}

// getTeamFilter get filter string for team
func getTeamFilter(teamName string) string {
	return fmt.Sprintf("OwningTeam/FullName eq '%s'", teamName)
}

// getProjectIdsFilter get filter string for project-id option
func getProjectIdsFilter(projectIds string) string {
	if matched, _ := regexp.MatchString(`^\d+$`, projectIds); matched {
		return fmt.Sprintf("Id eq %s", projectIds)
	}
	if matched, _ := regexp.MatchString(`^\d+(,\s?\d+)+$`, projectIds); matched {
		return fmt.Sprintf("Id in (%s)", projectIds)
	}
	if matched, _ := regexp.MatchString(`^\d+\s?-\s?\d+$`, projectIds); matched {
		ids := strings.Split(projectIds, "-")
		min, max := getMinMax(ids)
		return fmt.Sprintf("Id ge %d and Id le %d", min, max)
	}

	log.Warn().Msg("--project-id has wrong param. It should be like --project-id 1 or 1,3,8 or 1-3")
	return "Id gt 0"
}

// getMinMax get min and max values
func getMinMax(ids []string) (min, max int) {
	min, _ = strconv.Atoi(strings.Trim(ids[0], " "))
	max, _ = strconv.Atoi(strings.Trim(ids[1], " "))
	if min > max {
		min, max = max, min
	}

	return min, max
}
