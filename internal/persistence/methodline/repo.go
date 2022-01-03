package methodline

import (
	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	"github.com/pkg/errors"
)

type Repo struct {
	soapClient soap.Adapter
}

func NewRepo(soapClient soap.Adapter) *Repo {
	return &Repo{soapClient: soapClient}
}

func (e *Repo) GetMethodLines(scanID, queryID, pathID string) ([]string, error) {
	resultPaths, resultPathErr := e.soapClient.GetResultPathsForQuery(scanID, queryID)
	if resultPathErr != nil {
		return nil, errors.Wrap(resultPathErr, "could not get result paths")
	}
	var output []string
	for _, resultPath := range resultPaths.GetResultPathsForQueryResult.Paths.Paths {
		if resultPath.PathID == pathID {
			for _, v := range resultPath.Node.Nodes {
				output = append(output, v.MethodLine)
			}
		}
	}
	return output, nil
}

func (e *Repo) GetMethodLinesByPath(scanID, queryID string) ([]*interfaces.ResultPath, error) {
	resultPaths, resultPathErr := e.soapClient.GetResultPathsForQuery(scanID, queryID)
	if resultPathErr != nil {
		return nil, errors.Wrap(resultPathErr, "could not get result paths")
	}
	var output []*interfaces.ResultPath
	for _, resultPath := range resultPaths.GetResultPathsForQueryResult.Paths.Paths {
		methodLine := interfaces.ResultPath{
			PathID:      resultPath.PathID,
			MethodLines: []string{},
		}
		for _, v := range resultPath.Node.Nodes {
			methodLine.MethodLines = append(methodLine.MethodLines, v.MethodLine)
		}
		output = append(output, &methodLine)
	}
	return output, nil
}
