package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/checkmarxDev/ast-sast-export/internal/ast"
	"github.com/checkmarxDev/ast-sast-export/internal/sast"
	"github.com/checkmarxDev/ast-sast-export/internal/soap"
	"github.com/pkg/errors"
)

type MetadataProvider interface {
	GetMetadataForQueryAndResult(scanID string, query *MetadataQuery, result *MetadataResult) (*MetadataRecord, error)
}

type MetadataSource struct {
	astQueryIDProvider   ast.QueryIDProvider
	similarityIDProvider sast.SimilarityIDProvider
	soapAdapter          soap.Adapter
	tmpDir               string
}

func NewMetadataSource(
	astQueryIDProvider ast.QueryIDProvider,
	similarityIDProvider sast.SimilarityIDProvider,
	soapAdapter soap.Adapter,
	tmpDir string,
) *MetadataSource {
	return &MetadataSource{
		astQueryIDProvider,
		similarityIDProvider,
		soapAdapter,
		tmpDir,
	}
}

func (e *MetadataSource) GetMetadataForQueryAndResult(
	scanID string, query *MetadataQuery, result *MetadataResult,
) (*MetadataRecord, error) {
	astQueryID, astQueryIDErr := e.astQueryIDProvider.GetQueryID(query.Language, query.Name, query.Group)
	if astQueryIDErr != nil {
		return nil, errors.Wrap(astQueryIDErr, "could not get AST query id")
	}
	resultPaths, resultPathErr := e.soapAdapter.GetResultPathsForQuery(scanID, query.QueryID)
	if resultPathErr != nil {
		return nil, errors.Wrap(resultPathErr, "could not get result paths")
	}
	var firstMethodLine, lastMethodLine string
	for _, resultPath := range resultPaths.GetResultPathsForQueryResult.Paths {
		if resultPath.PathID == result.PathID {
			firstMethodLine = resultPath.Nodes[0].MethodLine
			lastMethodLine = resultPath.Nodes[len(resultPath.Nodes)-1].MethodLine
		}
	}
	if firstMethodLine == "" || lastMethodLine == "" {
		return nil, fmt.Errorf("could not get method lines")
	}
	sourceResponse, sourceErr := e.soapAdapter.GetSourcesByScanID(scanID, []string{result.FirstNode.FileName, result.LastNode.FileName})
	if sourceErr != nil {
		return nil, errors.Wrap(sourceErr, "could not get source code")
	}
	sourceContent := sourceResponse.GetSourcesByScanIDResult.CxWSResponseSourcesContent
	firstFileName := filepath.Join(e.tmpDir, result.FirstNode.FileName)
	firstFileSource := sourceContent[0].CxWSResponseSourceContent.Source
	lastFileName := filepath.Join(e.tmpDir, result.LastNode.FileName)
	lastFileSource := sourceContent[1].CxWSResponseSourceContent.Source
	firstFileErr := createFileAndPath(firstFileName, []byte(firstFileSource), 0600, 0700)
	if firstFileErr != nil {
		return nil, errors.Wrap(sourceErr, "could not create first file")
	}
	lastFileErr := createFileAndPath(lastFileName, []byte(lastFileSource), 0600, 0700)
	if lastFileErr != nil {
		return nil, errors.Wrap(sourceErr, "could not create last file")
	}
	similarityID, similarityIDErr := e.similarityIDProvider.Calculate(
		firstFileName, result.FirstNode.Name, result.FirstNode.Line, result.FirstNode.Column, firstMethodLine,
		lastFileName, result.LastNode.Name, result.LastNode.Line, result.LastNode.Column, lastMethodLine,
		astQueryID,
	)
	if similarityIDErr != nil {
		return nil, errors.Wrap(sourceErr, "could not calculate similarity id")
	}
	return &MetadataRecord{
		QueryID:      query.QueryID,
		SimilarityID: similarityID,
		PathID:       result.PathID,
		ResultID:     result.ResultID,
	}, nil
}

func createFileAndPath(filename string, content []byte, filePerm, dirPerm os.FileMode) error {
	pathErr := os.MkdirAll(filepath.Dir(filename), dirPerm)
	if pathErr != nil {
		return pathErr
	}
	return os.WriteFile(filename, content, filePerm)
}
