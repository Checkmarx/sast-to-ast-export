package worker

import "runtime"

const (
	maxCPUs = 4
)

func GetNumCPU() int {
	numCPU := runtime.NumCPU() - 1
	// Not allow more than 4 cpu's
	if numCPU > maxCPUs {
		numCPU = maxCPUs
	}
	return numCPU
}
