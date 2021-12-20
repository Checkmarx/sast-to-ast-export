package database

type ComponentConfiguration struct {
	ID    int    `gorm:"primaryKey;column:Id"`
	Key   string `gorm:"primaryKey;column:Key"`
	Value string `gorm:"primaryKey;column:Value"`
}

type NodeResult struct {
	ResultID   int    `gorm:"primaryKey;autoIncrement:false;column:ResultId"`
	PathID     int    `gorm:"primaryKey;autoIncrement:false;column:Path_Id"`
	NodeID     int    `gorm:"primaryKey;autoIncrement:false;column:Node_Id"`
	FullName   string `gorm:"column:Full_Name"`
	ShortName  string `gorm:"column:Short_Name"`
	FileName   string `gorm:"column:File_Name"`
	Line       int    `gorm:"column:Line"`
	Col        int    `gorm:"column:Col"`
	Length     int    `gorm:"column:Length"`
	DomID      int    `gorm:"column:DOM_Id"`
	MethodLine int    `gorm:"column:Method_Line"`
}

type TaskScan struct {
	ID          int    `gorm:"column:Id"`
	ProjectID   int    `gorm:"column:ProjectId"`
	VersionDate string `gorm:"column:VersionDate"`
	SourceID    string `gorm:"column:SourceId"`
}
