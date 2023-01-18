package soap

import "encoding/xml"

type (
	// SOAP types

	Envelope struct {
		XMLName struct{} `xml:"Envelope"`
		Header  Header
		Body    Body
	}

	Header struct {
		XMLName  struct{} `xml:"Header"`
		Contents []byte   `xml:",innerxml"`
	}

	Body struct {
		XMLName  struct{} `xml:"Body"`
		Contents []byte   `xml:",innerxml"`
	}

	// GetResultPathsForQuery request types

	GetResultPathsForQueryRequest struct {
		XMLName xml.Name `xml:"chec:GetResultPathsForQuery"`
		ScanID  string   `xml:"chec:scanId"`
		QueryID string   `xml:"chec:queryId"`
	}

	GetResultPathsForQueryResponse struct {
		XMLName                      xml.Name `xml:"GetResultPathsForQueryResponse"`
		GetResultPathsForQueryResult GetResultPathsForQueryResult
	}

	GetResultPathsForQueryResult struct {
		XMLName      xml.Name `xml:"GetResultPathsForQueryResult"`
		IsSuccessful bool     `xml:"IsSuccesfull"`
		ErrorMessage string   `xml:"ErrorMessage"`
		Paths        Paths    `xml:"Paths"`
	}

	Paths struct {
		XMLName xml.Name     `xml:"Paths"`
		Paths   []ResultPath `xml:"CxWSResultPath"`
	}

	ResultPath struct {
		XMLName xml.Name `xml:"CxWSResultPath"`
		PathID  string   `xml:"PathId"`
		Node    Node     `xml:"Nodes"`
	}

	Node struct {
		XMLName xml.Name         `xml:"Nodes"`
		Nodes   []ResultPathNode `xml:"CxWSPathNode"`
	}

	ResultPathNode struct {
		XMLName    xml.Name `xml:"CxWSPathNode"`
		MethodLine string
	}

	// GetSourcesByScanID request types

	GetSourcesByScanIDRequest struct {
		XMLName         xml.Name                  `xml:"chec:GetSourcesByScanID"`
		ScanID          string                    `xml:"chec:scanID"`
		FilesToRetrieve GetSourcesFilesToRetrieve `xml:"chec:filesToRetreive"`
	}

	GetSourcesFilesToRetrieve struct {
		XMLName xml.Name `xml:"chec:filesToRetreive"`
		Strings []string `xml:"chec:string"`
	}

	GetSourcesByScanIDResponse struct {
		XMLName                  xml.Name `xml:"GetSourcesByScanIDResponse"`
		GetSourcesByScanIDResult GetSourcesByScanIDResult
	}

	GetSourcesByScanIDResult struct {
		XMLName                    xml.Name                   `xml:"GetSourcesByScanIDResult"`
		IsSuccessful               bool                       `xml:"IsSuccesfull"`
		ErrorMessage               string                     `xml:"ErrorMessage"`
		CxWSResponseSourcesContent CxWSResponseSourcesContent `xml:"cxWSResponseSourcesContent"`
	}

	CxWSResponseSourcesContent struct {
		XMLName                    xml.Name                    `xml:"cxWSResponseSourcesContent"`
		CxWSResponseSourceContents []CxWSResponseSourceContent `xml:"CxWSResponseSourceContent"`
	}

	CxWSResponseSourceContent struct {
		XMLName xml.Name `xml:"CxWSResponseSourceContent"`
		Source  string   `xml:"Source"`
	}

	// GetQueryCollection request types

	GetQueryCollectionRequest struct {
		XMLName xml.Name `xml:"chec:GetQueryCollection"`
	}

	GetQueryCollectionResponse struct {
		XMLName                  xml.Name                 `xml:"GetQueryCollectionResponse"`
		GetQueryCollectionResult GetQueryCollectionResult `xml:"GetQueryCollectionResult"`
	}

	GetQueryCollectionResult struct {
		XMLName      xml.Name    `xml:"GetQueryCollectionResult"`
		IsSuccessful bool        `xml:"IsSuccesfull"`
		QueryGroups  QueryGroups `xml:"QueryGroups"`
	}

	QueryGroups struct {
		XMLName        xml.Name         `xml:"QueryGroups"`
		CxWSQueryGroup []CxWSQueryGroup `xml:"CxWSQueryGroup"`
	}

	CxWSQueryGroup struct {
		XMLName           xml.Name `xml:"CxWSQueryGroup"`
		Name              string   `xml:"Name"`
		PackageID         int      `xml:"PackageId"`
		Queries           Queries  `xml:"Queries"`
		IsReadOnly        bool     `xml:"IsReadOnly"`
		IsEncrypted       bool     `xml:"IsEncrypted"`
		Description       string   `xml:"Description"`
		Language          int      `xml:"Language"`
		LanguageName      string   `xml:"LanguageName"`
		PackageTypeName   string   `xml:"PackageTypeName"`
		ProjectID         int      `xml:"ProjectId"`
		PackageType       string   `xml:"PackageType"`
		PackageFullName   string   `xml:"PackageFullName"`
		OwningTeam        int      `xml:"OwningTeam"`
		Status            string   `xml:"Status"`
		LanguageStateHash int64    `xml:"LanguageStateHash"`
		LanguageStateDate string   `xml:"LanguageStateDate"`
	}

	Queries struct {
		XMLName   xml.Name    `xml:"Queries"`
		CxWSQuery []CxWSQuery `xml:"CxWSQuery"`
	}

	CxWSQuery struct {
		XMLName          xml.Name   `xml:"CxWSQuery"`
		Name             string     `xml:"Name"`
		QueryID          int        `xml:"QueryId"`
		Source           string     `xml:"Source"`
		Cwe              int        `xml:"Cwe"`
		IsExecutable     bool       `xml:"IsExecutable"`
		IsEncrypted      bool       `xml:"IsEncrypted"`
		Severity         int        `xml:"Severity"`
		PackageID        int        `xml:"PackageId"`
		Status           string     `xml:"Status"`
		Type             string     `xml:"Type"`
		Categories       Categories `xml:"Categories"`
		CxDescriptionID  int        `xml:"CxDescriptionID"`
		QueryVersionCode int        `xml:"QueryVersionCode"`
		EngineMetadata   string     `xml:"EngineMetadata"`
	}

	Categories struct {
		XMLName         xml.Name          `xml:"Categories"`
		CxQueryCategory []CxQueryCategory `xml:"CxQueryCategory"`
	}

	CxQueryCategory struct {
		XMLName      xml.Name `xml:"CxQueryCategory"`
		ID           int      `xml:"Id"`
		CategoryName string   `xml:"CategoryName"`
		CategoryType int      `xml:"CategoryType"`
	}

	CategoryType struct {
		XMLName xml.Name `xml:"CategoryType"`
		ID      int      `xml:"Id"`
		Name    string   `xml:"Name"`
		Order   int      `xml:"Order"`
	}

	// GetPresetDetails request types

	GetPresetDetailsRequest struct {
		XMLName xml.Name `xml:"chec:GetPresetDetails"`
		ID      int      `xml:"chec:id"`
	}

	GetPresetDetailsResponse struct {
		XMLName                xml.Name               `xml:"GetPresetDetailsResponse"`
		GetPresetDetailsResult GetPresetDetailsResult `xml:"GetPresetDetailsResult"`
	}

	GetPresetDetailsResult struct {
		XMLName      xml.Name `xml:"GetPresetDetailsResult"`
		IsSuccessful bool     `xml:"IsSuccesfull"`
		Preset       Preset   `xml:"preset"`
	}

	Preset struct {
		XMLName             xml.Name `xml:"preset"`
		QueryIds            QueryIds `xml:"queryIds"`
		ID                  int      `xml:"id"`
		Name                string   `xml:"name"`
		OwningTeam          int      `xml:"owningteam"`
		IsPublic            bool     `xml:"isPublic"`
		Owner               string   `xml:"owner"`
		IsUserAllowToUpdate bool     `xml:"isUserAllowToUpdate"`
		IsUserAllowToDelete bool     `xml:"isUserAllowToDelete"`
		IsDuplicate         bool     `xml:"IsDuplicate"`
	}

	QueryIds struct {
		XMLName xml.Name `xml:"queryIds"`
		Long    []int    `xml:"long"`
	}

	// GetInstallationSettings request types

	GetInstallationSettingsRequest struct {
		XMLName xml.Name `xml:"chec:GetInstallationSettings"`
	}

	InstallationSetting []struct {
		XMLName         xml.Name `xml:"InstallationSetting"`
		Text            string   `xml:",chardata"`
		Name            string   `xml:"Name"`
		ID              string   `xml:"ID"`
		DNSName         string   `xml:"DNSName"`
		IP              string   `xml:"IP"`
		State           string   `xml:"State"`
		Version         string   `xml:"Version"`
		Hotfix          string   `xml:"Hotfix"`
		InstllationPath string   `xml:"InstllationPath"`
		IsInstalled     string   `xml:"IsInstalled"`
	}

	InstallationSettingsList struct {
		XMLName             xml.Name            `xml:"InstallationSettingsList"`
		InstallationSetting InstallationSetting `xml:"InstallationSetting"`
	}

	GetInstallationSettingsResult struct {
		XMLName                  xml.Name                 `xml:"GetInstallationSettingsResult"`
		IsSuccesfull             string                   `xml:"IsSuccesfull"`
		InstallationSettingsList InstallationSettingsList `xml:"InstallationSettingsList"`
	}

	GetInstallationSettingsResponse struct {
		Text                          string                        `xml:",chardata"`
		Xmlns                         string                        `xml:"xmlns,attr"`
		GetInstallationSettingsResult GetInstallationSettingsResult `xml:"GetInstallationSettingsResult"`
	}

	InstallationMapping struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Hotfix  string `json:"hotFix"`
	}
)
