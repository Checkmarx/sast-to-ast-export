package logging

import (
	"fmt"
	"github.com/checkmarxDev/ast-sast-export/internal/export"
	"os"
	"time"
)

// NewFileWriter creates a new log file, and returns it's descriptor
func NewFileWriter(serviceName string) (*os.File, error) {
	now := time.Now()
	logFileName := fmt.Sprintf("%s-%s.log", serviceName, now.Format(export.DateTimeFormat))
	return os.Create(logFileName)
}
