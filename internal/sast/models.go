package sast

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
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

type ReportRequest struct {
	ReportType string `json:"reportType"`
	ScanID     int    `json:"scanId"`
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
