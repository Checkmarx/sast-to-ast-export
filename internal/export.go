package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

type Export struct {
	Users    []User    `json:"users"`
	Roles    []Role    `json:"roles"`
	Projects []Project `json:"groups"`
}

func (c *Export) SaveToFile(basePath, prefix string) (string, error) {
	fileName := createFileName(prefix, "json")
	jsonData, err := json.Marshal(c)
	if err != nil {
		return fileName, err
	}
	filePath := filepath.Join(basePath, fileName)
	err = ioutil.WriteFile(filePath, jsonData, 0600)
	return fileName, err
}

func createFileName(prefix, extension string) string {
	currentTime := time.Now()
	return fmt.Sprintf("%s-%s.%s", prefix, currentTime.Format("2006-01-02-15-04-05"), extension)
}
