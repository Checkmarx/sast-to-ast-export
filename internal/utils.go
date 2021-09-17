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

func getLastScanByProject(list []LastTriagedScanProducer, projectID int) int {
	lastScan := 0
	for _, scan := range list {
		if scan.ScanID > lastScan && scan.ProjectID == projectID {
			lastScan = scan.ScanID
		}
	}
	return lastScan
}

func isScanInList(projectID, scanID int, list []LastTriagedScanProducer) bool {
	for _, a := range list {
		if a.ProjectID == projectID && a.ScanID == scanID {
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
	numCPU := runtime.NumCPU() - 1
	// Not allow more than 4 cpu's
	if numCPU > maxCPUs {
		numCPU = maxCPUs
	}
	return numCPU
}

func GetEncodingURL(params, str string) string {
	return url.QueryEscape(fmt.Sprintf(params, str))
}
