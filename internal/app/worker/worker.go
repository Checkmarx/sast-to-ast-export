package worker

import "runtime"

const (
	maxCPUs      = 4
	maxIOWorkers = 20
)

func GetNumCPU() int {
	numCPU := runtime.NumCPU() - 1
	// Not allow more than 4 cpu's
	if numCPU > maxCPUs {
		numCPU = maxCPUs
	}
	return numCPU
}

// GetIOWorkerCount returns a worker count suitable for I/O-bound concurrent HTTP calls.
// Unlike GetNumCPU, this is not tied to CPU count since goroutines waiting on network
// responses don't consume CPU. Capped at maxIOWorkers to avoid overwhelming the SAST server.
func GetIOWorkerCount() int {
	n := runtime.NumCPU() * 4
	if n > maxIOWorkers {
		n = maxIOWorkers
	}
	return n
}
