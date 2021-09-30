package internal

import (
	"fmt"
	"runtime"
	"time"
)

const (
	maxCPUs = 4
)

func GetDateFromDays(numDays int, now time.Time) string {
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
