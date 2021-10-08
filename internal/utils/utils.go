package utils

import "time"

type Retry struct {
	Attempts int
	MinSleep,
	MaxSleep time.Duration
}
