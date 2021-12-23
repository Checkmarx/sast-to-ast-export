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
	GetMetadataRecords(scanID string, query *Query) ([]*Record, error)
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

func (e *MetadataFactory) GetMetadataRecords(scanID string, query *Query) ([]*Record, error) {
	astQueryID, astQueryIDErr := e.astQueryIDProvider.GetQueryID(query.Language, query.Name, query.Group)
	if astQueryIDErr != nil {
		return nil, errors.Wrap(astQueryIDErr, "could not get AST query id")
	}
	methodLinesByPath, methodLineErr := e.methodLineProvider.GetMethodLinesByPath(scanID, query.QueryID)
	if methodLineErr != nil {
		return nil, errors.Wrap(methodLineErr, "could not get method lines")
	}
	filesToDownload := map[string]string{}
	for _, result := range query.Results {
		if _, ok1 := filesToDownload[result.FirstNode.FileName]; !ok1 {
			filesToDownload[result.FirstNode.FileName] = filepath.Join(e.tmpDir, result.FirstNode.FileName)
		}
		if _, ok2 := filesToDownload[result.LastNode.FileName]; !ok2 {
			filesToDownload[result.LastNode.FileName] = filepath.Join(e.tmpDir, result.LastNode.FileName)
		}
	}
	downloadErr := e.sourceProvider.DownloadSourceFiles(scanID, filesToDownload)
	if downloadErr != nil {
		return nil, errors.Wrap(downloadErr, "could not download source code")
	}
	var output []*Record
	for _, result := range query.Results {
		firstFileName := filesToDownload[result.FirstNode.FileName]
		lastFileName := filesToDownload[result.LastNode.FileName]
		methodLines := methodLinesByPath[result.PathID]
		similarityID, similarityIDErr := e.similarityIDProvider.Calculate(
			firstFileName, result.FirstNode.Name, result.FirstNode.Line, result.FirstNode.Column, methodLines[0],
			lastFileName, result.LastNode.Name, result.LastNode.Line, result.LastNode.Column, methodLines[len(methodLines)-1],
			astQueryID,
		)
		if similarityIDErr != nil {
			return nil, errors.Wrap(similarityIDErr, "could not calculate similarity id")
		}
		output = append(output, &Record{
			QueryID:      query.QueryID,
			SimilarityID: similarityID,
			PathID:       result.PathID,
			ResultID:     result.ResultID,
		})
	}
	return output, nil
}
