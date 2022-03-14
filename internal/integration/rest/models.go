package rest

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

	Team struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		FullName string `json:"fullName"`
		ParendID int    `json:"parentId"`
	}

	User struct {
		ID                       int      `json:"id"`
		UserName                 string   `json:"userName"`
		LastLoginDate            string   `json:"lastLoginDate"`
		RoleIDs                  []int    `json:"roleIds"`
		TeamIDs                  []int    `json:"teamIds"`
		AuthenticationProviderID int      `json:"authenticationProviderId"`
		CreationDate             string   `json:"creationDate"`
		FirstName                string   `json:"firstName"`
		LastName                 string   `json:"lastName"`
		Email                    string   `json:"email"`
		PhoneNumber              string   `json:"phoneNumber"`
		CellPhoneNumber          string   `json:"cellPhoneNumber"`
		JobTitle                 string   `json:"jobTitle"`
		Other                    string   `json:"other"`
		Country                  string   `json:"country"`
		Active                   bool     `json:"active"`
		ExpirationDate           string   `json:"expirationDate"`
		AllowedIPList            []string `json:"allowedIpList"`
		LocaleID                 int      `json:"localeId"`
	}

	SamlTeamMapping struct {
		ID                     int    `json:"id"`
		SamlIdentityProviderID int    `json:"samlIdentityProviderId"`
		TeamID                 int    `json:"teamId"`
		TeamFullPath           string `json:"teamFullPath"`
		SamlAttributeValue     string `json:"samlAttributeValue"`
	}
)
