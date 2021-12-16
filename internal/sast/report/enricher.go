package report

import "encoding/xml"

const (
	xmlHeader = `<?xml version="1.0" encoding="utf-8"?>`
)

type Enricher interface {
	AddSimilarity()
	Marshal() ([]byte, error)
}

type Report struct {
	report *CxXMLResults
}

func NewReport(reportData []byte) (*Report, error) {
	e := Report{}
	err := xml.Unmarshal(reportData, &e.report)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (e *Report) AddSimilarity() error {
	for i := 0; i < len(e.report.Queries); i++ {
		for j := 0; j < len(e.report.Queries[i].Results); j++ {
			for k := 0; k < len(e.report.Queries[i].Results[j].Paths); k++ {
				e.report.Queries[i].Results[j].Paths[k].NewSimilarityID = "1234"
			}
		}
	}
	return nil
}

func (e *Report) Marshal() ([]byte, error) {
	data, err := xml.MarshalIndent(e.report, "", "    ")
	if err != nil {
		return data, err
	}
	data = []byte(xmlHeader + "\n" + string(data))
	return data, nil
}
