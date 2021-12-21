package soap

import "encoding/xml"

type (
	GetSourcesByScanIDRequest struct {
		XMLName         xml.Name `xml:"GetSourcesByScanIDRequest"`
		ScanID          string   `xml:"scanID"`
		FilesToRetrieve []string `xml:"filesToRetreive"`
	}

	GetSourcesByScanIDResponse struct {
		XMLName                  xml.Name `xml:"GetSourcesByScanIDResponse"`
		GetSourcesByScanIDResult GetSourcesByScanIDResult
	}

	GetSourcesByScanIDResult struct {
		XMLName                    xml.Name `xml:"GetSourcesByScanIDResult"`
		IsSuccessful               bool
		CxWSResponseSourcesContent []CxWSResponseSourcesContent `xml:"cxWSResponseSourcesContent"`
	}

	CxWSResponseSourcesContent struct {
		XMLName                   xml.Name `xml:"cxWSResponseSourcesContent"`
		CxWSResponseSourceContent CxWSResponseSourceContent
	}

	CxWSResponseSourceContent struct {
		XMLName      xml.Name `xml:"CxWSResponseSourceContent"`
		IsSuccessful bool     `xml:"IsSuccesfull"`
		Source       string
	}

	GetResultPathsForQueryRequest struct {
		XMLName xml.Name `xml:"GetResultPathsForQueryRequest"`
		ScanID  string   `xml:"scanId"`
		QueryID string   `xml:"queryId"`
	}

	GetResultPathsForQueryResponse struct {
		XMLName                      xml.Name `xml:"GetResultPathsForQueryResponse"`
		GetResultPathsForQueryResult GetResultPathsForQueryResult
	}

	GetResultPathsForQueryResult struct {
		XMLName      xml.Name `xml:"GetResultPathsForQueryResult"`
		IsSuccessful bool     `xml:"IsSuccesfull"`
		Paths        []ResultPath
	}

	ResultPath struct {
		XMLName xml.Name `xml:"CxWSResultPath"`
		PathID  string   `xml:"pathId"`
		Nodes   []ResultPathNode
	}

	ResultPathNode struct {
		XMLName    xml.Name `xml:"CxWSPathNode"`
		MethodLine string
	}
)
