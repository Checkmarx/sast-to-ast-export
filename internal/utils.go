package internal

import (
	"fmt"
	"net/url"
	"runtime"
	"time"
)

func dataToArray(data LastTriagedResponse) []LastTriagedScanProducer {
	if isDebug {
		fmt.Printf("dataToArray len: %d\n", len(data.Value))
	}
	var result []LastTriagedScanProducer
	for _, v := range data.Value {
		result = append(result, LastTriagedScanProducer{
			ProjectID: v.Scan.ProjectID,
			ScanID:    v.ScanID,
		})
	}

	return GetUniqueAndMax(result)
}

func contains(projectId, scanId int, list []LastTriagedScanProducer) bool {
	for _, a := range list {
		if a.ProjectID == projectId && a.ScanID == scanId {
			return true
		}
	}
	return false
}

func GetUniqueAndMax(strList []LastTriagedScanProducer) []LastTriagedScanProducer {
	var result []LastTriagedScanProducer
	for _, item := range strList {
		if contains(item.ProjectID, item.ScanID, result) == false {
			max := GetMax(strList, item.ProjectID)
			if max > 0 && max == item.ScanID {
				result = append(result, item)
			}
		}
	}
	if isDebug {
		fmt.Printf("result: %v\n", result)
		fmt.Printf("result len: %d\n", len(result))
	}

	return result
}

func GetMax(list []LastTriagedScanProducer, projectId int) int {
	maxScan := 0
	for _, scan := range list {
		if scan.ScanID > maxScan && scan.ProjectID == projectId {
			maxScan = scan.ScanID
		}
	}

	return maxScan
}

func GetDateFromDays(numDays int) string {
	now := time.Now()

	date := now.AddDate(0, 0, -numDays)

	return fmt.Sprintf("%d-%d-%d", date.Year(), date.Month(), date.Day())
}

func GetNumCPU() int {
	numCpu := runtime.NumCPU() - 1
	// Not allow more than 4 cpu's
	if numCpu > 4 {
		numCpu = 4
	}
	if isDebug {
		fmt.Printf("NumCPU used: %v\n", numCpu)
	}
	return numCpu
}

func GetEncodingUrl(params, str string) string {
	return url.QueryEscape(fmt.Sprintf(params, str))
}
