package internal

import (
	"fmt"
	"net/url"
	"runtime"
	"time"
)

const (
	maxCPUs = 4
)

func convertTriagedScansResponseToLastScansList(triagedScansResponse LastTriagedResponse) []LastTriagedScanProducer {
	var result []LastTriagedScanProducer
	for _, v := range triagedScansResponse.Value {
		result = append(result, LastTriagedScanProducer{
			ProjectID: v.Scan.ProjectID,
			ScanID:    v.ScanID,
		})
	}
	return getLastScansByProject(result)
}

func getLastScansByProject(scans []LastTriagedScanProducer) []LastTriagedScanProducer {
	var result []LastTriagedScanProducer
	for _, item := range scans {
		if !isScanInList(item.ProjectID, item.ScanID, result) {
			lastScan := getLastScanByProject(scans, item.ProjectID)
			if lastScan > 0 && lastScan == item.ScanID {
				result = append(result, item)
			}
		}
	}
	return result
}

func getLastScanByProject(list []LastTriagedScanProducer, projectId int) int {
	lastScan := 0
	for _, scan := range list {
		if scan.ScanID > lastScan && scan.ProjectID == projectId {
			lastScan = scan.ScanID
		}
	}
	return lastScan
}

func isScanInList(projectId, scanId int, list []LastTriagedScanProducer) bool {
	for _, a := range list {
		if a.ProjectID == projectId && a.ScanID == scanId {
			return true
		}
	}
	return false
}

func GetDateFromDays(numDays int) string {
	now := time.Now()

	date := now.AddDate(0, 0, -numDays)

	return fmt.Sprintf("%d-%d-%d", date.Year(), date.Month(), date.Day())
}

func GetNumCPU() int {
	numCpu := runtime.NumCPU() - 1
	// Not allow more than 4 cpu's
	if numCpu > maxCPUs {
		numCpu = maxCPUs
	}
	return numCpu
}

func GetEncodingURL(params, str string) string {
	return url.QueryEscape(fmt.Sprintf(params, str))
}
