package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal"
)

// NewFileWriter creates a new log file, and returns it's descriptor
func NewFileWriter(serviceName string) (*os.File, error) {
	now := time.Now()
	logFileName := fmt.Sprintf("%s-%s.log", serviceName, now.Format(internal.DateTimeFormat))
	return os.Create(logFileName)
}
