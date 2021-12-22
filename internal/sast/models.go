package sast

type (
	AccessToken struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	StatusResponse struct {
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

	ReportResponse struct {
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

	ReportRequest struct {
		ReportType string `json:"reportType"`
		ScanID     int    `json:"scanId"`
	}

	ODataProjectsWithLastScanID struct {
		OdataContext string                  `json:"@odata.context"`
		Value        []ProjectWithLastScanID `json:"value"`
	}

	ProjectWithLastScanID struct {
		ID         int `json:"Id"`
		LastScanID int `json:"LastScanId"`
	}

	ODataTriagedResultsByScan struct {
		Value []TriagedScanResult
	}

	TriagedScanResult struct {
		ID int `json:"Id"`
	}
)
