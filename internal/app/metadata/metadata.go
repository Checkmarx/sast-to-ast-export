package metadata

import (
	"path/filepath"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/similarity"
	"github.com/checkmarxDev/ast-sast-export/internal/persistence/ast_query_id"
	"github.com/checkmarxDev/ast-sast-export/internal/persistence/method_line"
	"github.com/checkmarxDev/ast-sast-export/internal/persistence/source"

	"github.com/pkg/errors"
)

type MetadataProvider interface {
	GetMetadataForQueryAndResult(scanID string, query *MetadataQuery, result *MetadataResult) (*MetadataRecord, error)
}

type MetadataFactory struct {
	astQueryIDProvider   ast_query_id.QueryIDProvider
	similarityIDProvider similarity.SimilarityIDProvider
	sourceProvider       source.SourceProvider
	methodLineProvider   method_line.Provider
	tmpDir               string
}

func NewMetadataFactory(
	astQueryIDProvider ast_query_id.QueryIDProvider,
	similarityIDProvider similarity.SimilarityIDProvider,
	sourceProvider source.SourceProvider,
	methodLineProvider method_line.Provider,
	tmpDir string,
) *MetadataFactory {
	return &MetadataFactory{
		astQueryIDProvider,
		similarityIDProvider,
		sourceProvider,
		methodLineProvider,
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
	methodLines, methodLineErr := e.methodLineProvider.GetMethodLines(scanID, query.QueryID, result.PathID)
	if methodLineErr != nil {
		return nil, errors.Wrap(methodLineErr, "could not get method lines")
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
		firstFileName, result.FirstNode.Name, result.FirstNode.Line, result.FirstNode.Column, methodLines[0],
		lastFileName, result.LastNode.Name, result.LastNode.Line, result.LastNode.Column, methodLines[len(methodLines)-1],
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
