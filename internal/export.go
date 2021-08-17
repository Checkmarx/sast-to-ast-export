package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"
)

type Export struct {
	FilePrefix string
	Data       ExportData
}

type ExportData struct {
	Users    []User    `json:"users"`
	Roles    []Role    `json:"roles"`
	Projects []Project `json:"projects"`
}

func (c *Export) WriteToFile(out io.Writer) error {
	jsonData, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}
	_, err = io.WriteString(out, string(jsonData))
	return err
}

func (c *Export) CreateFileName(basePath string) string {
	currentTime := time.Now()
	fileName := fmt.Sprintf("%s-%s.json", c.FilePrefix, currentTime.Format("2006-01-02-15-04-05"))
	return filepath.Join(basePath, fileName)
}
