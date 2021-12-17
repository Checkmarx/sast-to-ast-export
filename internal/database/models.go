package database

type ComponentConfiguration struct {
	ID    int
	Key   string
	Value string
}

type NodeResult struct {
	ResultID   int    `gorm:"ResultId"`
	PathID     int    `gorm:"Path_Id"`
	NodeID     int    `gorm:"Node_Id"`
	FullName   string `gorm:"Full_Name"`
	ShortName  string `gorm:"Short_Name"`
	FileName   string `gorm:"File_Name"`
	Line       int    `gorm:"Line"`
	Col        int    `gorm:"Col"`
	Length     int    `gorm:"Length"`
	DomID      int    `gorm:"DOM_Id"`
	MethodLine int    `gorm:"Method_Line"`
}

type TaskScan struct {
	ID          int    `gorm:"Id"`
	ProjectID   int    `gorm:"ProjectId"`
	VersionDate string `gorm:"VersionDate"`
	SourceID    string `gorm:"SourceId"`
}
