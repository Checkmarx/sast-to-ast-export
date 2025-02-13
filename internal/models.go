package internal

import "time"

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
	RunTime          time.Time
	QueryMappingFile string
	NestedTeams      bool
	SimIDVersion     int
	ExcludeFile      string
	ExcludeFiles     []string
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
	ProjectID             int
	EngineConfigurationID int
}

type PresetJob struct {
	PresetID int
}

type PresetConsumeOutput struct {
	Err      error
	PresetID int
}
