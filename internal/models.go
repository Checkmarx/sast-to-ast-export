package internal

import "time"

type Result []interface{}

type Args struct {
	URL,
	Username,
	Password,
	OutputPath,
	ProductName string
	Export              []string
	ProjectsActiveSince int
	Debug               bool
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type User struct {
	ID            int    `json:"id"`
	UserName      string `json:"userName"`
	LastLoginDate string `json:"lastLoginDate"`
	RoleIds       []int  `json:"roleIds"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Email         string `json:"email"`
}

type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Project struct {
	ID       int    `json:"id"`
	TeamID   int    `json:"teamId"`
	Name     string `json:"name"`
	IsPublic bool   `json:"isPublic"`
}

type StatusResponse struct {
	Link struct {
		Rel string `json:"rel"`
		URI string `json:"uri"`
	} `json:"link"`
	ContentType string `json:"contentType"`
	Status      struct {
		ID    int    `json:"id"`
		Value string `json:"value"`
	} `json:"status"`
}

type ReportResponse struct {
	ReportID int `json:"ReportId" groups:"out"`
	Links    struct {
		Report struct {
			Rel string `json:"rel"`
			URI string `json:"uri"`
		} `json:"ReportResponse"`
		Status struct {
			Rel string `json:"rel"`
			URI string `json:"uri"`
		} `json:"status"`
	} `json:"links"`
}

type ValueOdata struct {
	ID     int            `json:"Id"`
	ScanID int            `json:"ScanId"`
	Date   time.Time      `json:"Date"`
	Scan   ValueOdataScan `json:"Scan"`
}

type ValueOdataScan struct {
	ProjectID int `json:"ProjectId"`
}

type TriagedScansResponse struct {
	OdataContext string       `json:"@odata.context"`
	Value        []ValueOdata `json:"value"`
}

type ReportJob struct {
	ProjectID  int
	ScanID     int
	ReportType string
}

type ReportRequest struct {
	ReportType string `json:"reportType"`
	ScanID     int    `json:"scanId"`
}

type TriagedScan struct {
	ProjectID int
	ScanID    int
}

type ODataProjectsWithLastScanID struct {
	OdataContext string                  `json:"@odata.context"`
	Value        []ProjectWithLastScanID `json:"value"`
}

type ProjectWithLastScanID struct {
	ID         int `json:"Id"`
	LastScanID int `json:"LastScanId"`
}

type ODataTriagedResultsByScan struct {
	Value []TriagedScanResult
}

type TriagedScanResult struct {
	ID int `json:"Id"`
}
