package report

import (
	"encoding/xml"
	"path/filepath"
	"strconv"

	"github.com/checkmarxDev/ast-sast-export/internal/database/store"
	"github.com/checkmarxDev/ast-sast-export/internal/sast"
	"github.com/pkg/errors"
)

const (
	xmlHeader = `<?xml version="1.0" encoding="utf-8"?>`
)

type Enricher interface {
	AddSimilarity() error
	Parse(reportData []byte) error
	Marshal() ([]byte, error)
}

type Report struct {
	report               *CxXMLResults
	source               SourceProvider
	nodeResultsStore     store.NodeResultsStore
	similarityCalculator sast.SimilarityCalculator
}

func NewReport(source SourceProvider, nodeResultsStore store.NodeResultsStore, similarityCalculator sast.SimilarityCalculator) *Report {
	return &Report{source: source, nodeResultsStore: nodeResultsStore, similarityCalculator: similarityCalculator}
}

func (e *Report) Parse(reportData []byte) error {
	err := xml.Unmarshal(reportData, &e.report)
	if err != nil {
		return errors.Wrap(err, "could not parse report data")
	}
	return nil
}

func (e *Report) AddSimilarity() error {
	for i := 0; i < len(e.report.Queries); i++ {
		for j := 0; j < len(e.report.Queries[i].Results); j++ {
			for k := 0; k < len(e.report.Queries[i].Results[j].Paths); k++ {
				scan := e.report
				query := e.report.Queries[i]
				path := e.report.Queries[i].Results[j].Paths[k]
				pathNode1 := path.PathNodes[0]
				pathNode1Id := 1
				pathNode2 := path.PathNodes[0]
				pathNode2Id := 1
				if len(path.PathNodes) > 1 {
					pathNode2 = path.PathNodes[len(path.PathNodes)-1]
					pathNode2Id = len(path.PathNodes)
				}
				basePath, basePathErr := e.source.GetBasePath(scan.ScanID)
				if basePathErr != nil {
					return errors.Wrap(basePathErr, "could not get source base path")
				}
				filename1 := filepath.Join(basePath, pathNode1.FileName)
				filename2 := filepath.Join(basePath, pathNode2.FileName)
				resultPath1, resultPath1Err := e.nodeResultsStore.GetByResultPathAndNode(path.ResultID, path.PathID, pathNode1Id)
				if resultPath1Err != nil {
					return errors.Wrapf(resultPath1Err, "could not get result path #%d", pathNode1Id)
				}
				resultPath2, resultPath2Err := e.nodeResultsStore.GetByResultPathAndNode(path.ResultID, path.PathID, pathNode2Id)
				if resultPath2Err != nil {
					return errors.Wrapf(resultPath2Err, "could not get result path #%d", pathNode2Id)
				}
				similarityID, similarityIDErr := e.similarityCalculator.Calculate(
					filename1, pathNode1.Name, pathNode1.Line, pathNode1.Column, strconv.Itoa(resultPath1.MethodLine),
					filename2, pathNode2.Name, pathNode2.Line, pathNode2.Column, strconv.Itoa(resultPath2.MethodLine),
					query.ID,
				)
				if similarityIDErr != nil {
					return errors.Wrap(similarityIDErr, "could not calculate similarity id")
				}
				e.report.Queries[i].Results[j].Paths[k].NewSimilarityID = similarityID
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
