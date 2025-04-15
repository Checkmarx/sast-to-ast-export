package internal

import (
	"encoding/xml"
	"time"
)

type Args struct {
	URL,
	Username,
	Password,
	OutputPath,
	ProductName string
	Export                      []string
	ProjectsActiveSince         int
	IsDefaultProjectActiveSince bool
	Debug                       bool
	DBConnectionString,
	ProjectsIDs,
	TeamName string
	RunTime           time.Time
	QueryMappingFile  string
	QueryRenamingFile string

	NestedTeams      bool
	SimIDVersion     int
	ExcludeFile      string
	ExcludeFiles     []string
	CustomExtensions string
}

type ReportJob struct {
	ProjectID  int
	ScanID     int
	ReportType string
}

type TriagedScan struct {
	ProjectID int
	ScanID    int
}

type EngineConfig struct {
	ProjectID               int
	EngineConfigurationID   int
	EngineConfigurationName string
}

type EngineConfigMapping struct {
	EngineConfigurationID int    `json:"id"`
	Name                  string `json:"name"`
}

type EngineKey struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type Configuration struct {
	Name string      `json:"Name"`
	Keys interface{} `json:"Keys"`
}

type JoinedConfig struct {
	ProjectID               int         `json:"ProjectID"`
	EngineConfigurationID   int         `json:"EngineConfigurationID"`
	EngineConfigurationName string      `json:"Name"`
	ConfigurationKeys       []EngineKey `json:"Keys"`
}

type PresetJob struct {
	PresetID int
}

type PresetConsumeOutput struct {
	Err      error
	PresetID int
}

type EngineKeysData struct {
	EngineConfig struct {
		Configurations struct {
			Configuration []Configuration `json:"Configuration"`
		} `json:"Configurations"`
	} `json:"EngineConfig"`
}

type CustomExtensionsList struct {
	XMLName         xml.Name          `xml:"CustomExtensionsList"`
	CustomExtension []CustomExtension `xml:"CustomExtension"`
}

type CustomExtension struct {
	XMLName       xml.Name `xml:"CustomExtension"`
	Language      string   `xml:"Language"`
	Extension     string   `xml:"Extension"`
	LanguageGroup string   `xml:"LanguageGroup"`
}
