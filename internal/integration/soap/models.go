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
)
