package internal

import (
	"fmt"
	"os/exec"
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

// OpenPathInExplorer opens file explorer if running on Windows; does nothing on other platforms
func OpenPathInExplorer(path string) error {
	if runtime.GOOS == "windows" {
		cmdErr := exec.Command(`explorer`, path).Run() //nolint:gosec
		// ignore exit status 1, it was being returned even on success
		if cmdErr != nil && cmdErr.Error() != "exit status 1" {
			return fmt.Errorf("could not open temporary folder")
		}
	}
	return nil
}
