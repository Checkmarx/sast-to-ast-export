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

	ScanSettings struct {
		Project struct {
			ID int `json:"id"`
		} `json:"project"`
		EngineConfiguration struct {
			ID int `json:"id"`
		} `json:"engineConfiguration"`
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

	ODataProjectsWithCustomFields struct {
		Value []ProjectWithCustomFields `json:"value"`
	}

	ProjectWithCustomFields struct {
		ID           int            `json:"Id"`
		TeamID       int            `json:"OwningTeamId"`
		Name         string         `json:"Name"`
		IsPublic     bool           `json:"IsPublic"`
		PresetID     int            `json:"PresetId"`
		CreatedDate  string         `json:"CreatedDate"`
		CustomFields []*CustomField `json:"CustomFields"`
	}

	Project struct {
		ID            int            `json:"id"`
		TeamID        int            `json:"teamId"`
		Name          string         `json:"name"`
		IsPublic      bool           `json:"isPublic"`
		PresetID      int            `json:"presetId"`
		CreatedDate   string         `json:"createdDate"`
		Configuration *Configuration `json:"configuration"`
	}

	Configuration struct {
		CustomFields []*CustomField `json:"customFields"`
	}

	CustomField struct {
		FieldName  string `json:"fieldName"`
		FieldValue string `json:"fieldValue"`
	}

	ProjectOData struct {
		ID           int            `json:"Id"`
		CreatedDate  string         `json:"CreatedDate"`
		CustomFields []*CustomField `json:"CustomFields"`
	}

	PresetShort struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		OwnerName string `json:"ownerName"`
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

	StatusEngineServer struct {
		ID    int    `json:"id"`
		Value string `json:"value"`
	}

	LinkEngineServer struct {
		Rel string `json:"rel"`
		URI string `json:"uri"`
	}

	OfflineReasonCodeEngineServer struct {
		ID    int    `json:"id"`
		Value string `json:"value"`
	}

	EngineServer struct {
		ID                             int                           `json:"id"`
		Name                           string                        `json:"name"`
		URI                            string                        `json:"uri"`
		MinLoc                         int                           `json:"minLoc"`
		MaxLoc                         int                           `json:"maxLoc"`
		MaxScans                       int                           `json:"maxScans"`
		CxVersion                      string                        `json:"cxVersion"`
		OperatingSystem                string                        `json:"operatingSystem"`
		Status                         StatusEngineServer            `json:"status"`
		Link                           LinkEngineServer              `json:"link"`
		OfflineReasonCode              OfflineReasonCodeEngineServer `json:"offlineReasonCode"`
		OfflineReasonMessage           string                        `json:"offlineReasonMessage"`
		OfflineReasonMessageParameters string                        `json:"offlineReasonMessageParameters"`
	}
)
