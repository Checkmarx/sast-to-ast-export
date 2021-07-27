package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Export struct {
	Users    []User    `json:"users"`
	Roles    []Role    `json:"roles"`
	Projects []Project `json:"groups"`
}

func (c *Export) WriteToFile(file *os.File) error {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return err
	}
	_, err = file.Write(jsonData)
	return err
}

func (c *Export) CreateFileName(basePath, prefix string) string {
	currentTime := time.Now()
	fileName := fmt.Sprintf("%s-%s.json", prefix, currentTime.Format("2006-01-02-15-04-05"))
	return filepath.Join(basePath, fileName)
}
