package metadata

import (
	"fmt"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/similarity"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	"github.com/checkmarxDev/ast-sast-export/internal/persistence/ast_query_id"
	"github.com/checkmarxDev/ast-sast-export/internal/persistence/source"
	"path/filepath"

	"github.com/pkg/errors"
)

type MetadataProvider interface {
	GetMetadataForQueryAndResult(scanID string, query *MetadataQuery, result *MetadataResult) (*MetadataRecord, error)
}

type MetadataFactory struct {
	astQueryIDProvider   ast_query_id.QueryIDProvider
	similarityIDProvider similarity.SimilarityIDProvider
	soapAdapter          soap.Adapter
	sourceProvider       source.SourceProvider
	tmpDir               string
}

func NewMetadataFactory(
	astQueryIDProvider ast_query_id.QueryIDProvider,
	similarityIDProvider similarity.SimilarityIDProvider,
	soapAdapter soap.Adapter,
	sourceProvider source.SourceProvider,
	tmpDir string,
) *MetadataFactory {
	return &MetadataFactory{
		astQueryIDProvider,
		similarityIDProvider,
		soapAdapter,
		sourceProvider,
		tmpDir,
	}
}

func (e *MetadataFactory) GetMetadataForQueryAndResult(
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
	for _, resultPath := range resultPaths.GetResultPathsForQueryResult.Paths.Paths {
		if resultPath.PathID == result.PathID {
			firstMethodLine = resultPath.Node.Nodes[0].MethodLine
			lastMethodLine = resultPath.Node.Nodes[len(resultPath.Node.Nodes)-1].MethodLine
			break
		}
	}
	if firstMethodLine == "" || lastMethodLine == "" {
		return nil, fmt.Errorf("could not get method lines")
	}
	firstFileName := filepath.Join(e.tmpDir, result.FirstNode.FileName)
	lastFileName := filepath.Join(e.tmpDir, result.LastNode.FileName)
	filesToDownload := map[string]string{
		result.FirstNode.FileName: firstFileName,
		result.LastNode.FileName:  lastFileName,
	}
	downloadErr := e.sourceProvider.DownloadSourceFiles(scanID, filesToDownload)
	if downloadErr != nil {
		return nil, errors.Wrap(downloadErr, "could not download source code")
	}
	similarityID, similarityIDErr := e.similarityIDProvider.Calculate(
		firstFileName, result.FirstNode.Name, result.FirstNode.Line, result.FirstNode.Column, firstMethodLine,
		lastFileName, result.LastNode.Name, result.LastNode.Line, result.LastNode.Column, lastMethodLine,
		astQueryID,
	)
	if similarityIDErr != nil {
		return nil, errors.Wrap(similarityIDErr, "could not calculate similarity id")
	}
	return &MetadataRecord{
		QueryID:      query.QueryID,
		SimilarityID: similarityID,
		PathID:       result.PathID,
		ResultID:     result.ResultID,
	}, nil
}
