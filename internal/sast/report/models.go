package report

import "encoding/xml"

type (
	CxXMLResults struct {
		XMLName                  xml.Name `xml:"CxXMLResults"`
		InitiatorName            string   `xml:"InitiatorName,attr"`
		Owner                    string   `xml:"Owner,attr"`
		ScanID                   string   `xml:"ScanId,attr"`
		ProjectID                string   `xml:"ProjectId,attr"`
		ProjectName              string   `xml:"ProjectName,attr"`
		TeamFullPathOnReportDate string   `xml:"TeamFullPathOnReportDate,attr"`
		DeepLink                 string   `xml:"DeepLink,attr"`
		ScanStart                string   `xml:"ScanStart,attr"`
		Preset                   string   `xml:"Preset,attr"`
		ScanTime                 string   `xml:"ScanTime,attr"`
		LinesOfCodeScanned       string   `xml:"LinesOfCodeScanned,attr"`
		FilesScanned             string   `xml:"FilesScanned,attr"`
		ReportCreationTime       string   `xml:"ReportCreationTime,attr"`
		Team                     string   `xml:"Team,attr"`
		CheckmarxVersion         string   `xml:"CheckmarxVersion,attr"`
		ScanComments             string   `xml:"ScanComments,attr"`
		ScanType                 string   `xml:"ScanType,attr"`
		SourceOrigin             string   `xml:"SourceOrigin,attr"`
		Visibility               string   `xml:"Visibility,attr"`
		Queries                  []Query  `xml:"Query"`
	}

	Query struct {
		XMLName            xml.Name `xml:"Query"`
		ID                 string   `xml:"id,attr"`
		CweID              string   `xml:"cweId,attr"`
		Name               string   `xml:"name,attr"`
		Group              string   `xml:"group,attr"`
		Severity           string   `xml:"Severity,attr"`
		Language           string   `xml:"Language,attr"`
		LanguageHash       string   `xml:"LanguageHash,attr"`
		LanguageChangeDate string   `xml:"LanguageChangeDate,attr"`
		SeverityIndex      string   `xml:"SeverityIndex,attr"`
		QueryPath          string   `xml:"QueryPath,attr"`
		QueryVersionCode   string   `xml:"QueryVersionCode,attr"`
		Results            []Result `xml:"Result"`
	}

	Result struct {
		XMLName       xml.Name `xml:"Result"`
		NodeID        string   `xml:"NodeId,attr"`
		FileName      string   `xml:"FileName,attr"`
		Status        string   `xml:"Status,attr"`
		Line          string   `xml:"Line,attr"`
		Column        string   `xml:"Column,attr"`
		FalsePositive string   `xml:"FalsePositive,attr"`
		Severity      string   `xml:"Severity,attr"`
		AssignToUser  string   `xml:"AssignToUser,attr"`
		State         string   `xml:"state,attr"`
		Remark        string   `xml:"Remark,attr"`
		DeepLink      string   `xml:"DeepLink,attr"`
		SeverityIndex string   `xml:"SeverityIndex,attr"`
		DetectionDate string   `xml:"DetectionDate,attr"`
		Paths         []Path   `xml:"Path"`
	}

	Path struct {
		XMLName         xml.Name   `xml:"Path"`
		ResultID        string     `xml:"ResultId,attr"`
		PathID          string     `xml:"PathId,attr"`
		SimilarityID    string     `xml:"SimilarityId,attr"`
		NewSimilarityID string     `xml:"NewSimilarityId,attr"`
		PathNodes       []PathNode `xml:"PathNode"`
	}

	PathNode struct {
		XMLName  xml.Name `xml:"PathNode"`
		FileName string   `xml:"FileName"`
		Line     string   `xml:"Line"`
		Column   string   `xml:"Column"`
		NodeID   string   `xml:"NodeId"`
		Name     string   `xml:"Name"`
		Type     string   `xml:"Type"`
		Length   string   `xml:"Length"`
		Snippet  Snippet  `xml:"Snippet"`
	}

	Snippet struct {
		XMLName xml.Name `xml:"Snippet"`
		Line    Line     `xml:"Line"`
	}

	Line struct {
		XMLName xml.Name `xml:"Line"`
		Number  string   `xml:"Number"`
		Code    string   `xml:"Code"`
	}
)
