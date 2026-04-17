package metadata

import (
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/checkmarxDev/ast-sast-export/internal/app/report"
	"github.com/checkmarxDev/ast-sast-export/internal/app/worker"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/similarity"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type Provider interface {
	GetMetadataRecord(scanID string, queries []*Query) (*Record, error)
}

type Factory struct {
	astQueryIDProvider   interfaces.ASTQueryIDProvider
	similarityIDProvider similarity.IDProvider
	sourceProvider       interfaces.SourceFileRepo
	methodLineProvider   interfaces.MethodLineRepo
	tmpDir               string
	simIDVersion         int
	rmvDir               string
	customExtensions     string
}

func NewMetadataFactory(
	astQueryIDProvider interfaces.ASTQueryIDProvider,
	similarityIDProvider similarity.IDProvider,
	sourceProvider interfaces.SourceFileRepo,
	methodLineProvider interfaces.MethodLineRepo,
	tmpDir string,
	simIDVersion int,
	rmvDir string,
	customExtensions string,
) *Factory {
	return &Factory{
		astQueryIDProvider,
		similarityIDProvider,
		sourceProvider,
		methodLineProvider,
		tmpDir,
		simIDVersion,
		rmvDir,
		customExtensions,
	}
}

//nolint:funlen,gocyclo
func (e *Factory) GetMetadataRecord(scanID string, queries []*Query) (*Record, error) {
	output := &Record{Queries: []*RecordQuery{}}

	for queryIdx, query := range queries {
		output.Queries = append(output.Queries, &RecordQuery{QueryID: query.QueryID})
		astQueryID, astQueryIDErr := e.astQueryIDProvider.GetQueryID(query.Language, query.Name, query.Group, query.QueryID)
		if astQueryIDErr != nil {
			return nil, errors.Wrapf(
				astQueryIDErr,
				"could not get AST query id for language %s, group %s, and name %s",
				query.Language,
				query.Group,
				query.Name,
			)
		}
		methodLinesByPath, methodLineErr := e.methodLineProvider.GetMethodLinesByPath(scanID, query.QueryID)
		if methodLineErr != nil {
			return nil, errors.Wrap(methodLineErr, "could not get method lines")
		}
		var filesToDownload []interfaces.SourceFile
		fileMap := make(map[string]interfaces.SourceFile)
		for _, result := range query.Results {
			firstFile := filepath.Join(result.ResultID, result.FirstNode.FileName)
			lastFile := filepath.Join(result.ResultID, result.LastNode.FileName)

			if ok1 := findSourceFile(result.ResultID, firstFile, filesToDownload); ok1 == nil {
				sf := interfaces.SourceFile{
					ResultID:   result.ResultID,
					RemoteName: result.FirstNode.FileName,
					LocalName:  filepath.Join(e.tmpDir, result.ResultID, result.FirstNode.FileName),
				}
				filesToDownload = append(filesToDownload, sf)
				fileMap[result.ResultID+"|"+result.FirstNode.FileName] = sf
			}

			if ok2 := findSourceFile(result.ResultID, lastFile, filesToDownload); ok2 == nil {
				sf := interfaces.SourceFile{
					ResultID:   result.ResultID,
					RemoteName: result.LastNode.FileName,
					LocalName:  filepath.Join(e.tmpDir, result.ResultID, result.LastNode.FileName),
				}
				filesToDownload = append(filesToDownload, sf)
				fileMap[result.ResultID+"|"+result.LastNode.FileName] = sf
			}
		}
		downloadErr := e.sourceProvider.DownloadSourceFiles(scanID, filesToDownload, e.rmvDir)
		if downloadErr != nil {
			return nil, errors.Wrap(downloadErr, "could not download source code")
		}

		// produce calculation jobs
		similarityCalculationJobs := make(chan SimilarityCalculationJob)
		q := query
		go func() {
			for _, result := range q.Results {
				firstSourceFile := fileMap[result.ResultID+"|"+result.FirstNode.FileName]
				lastSourceFile := fileMap[result.ResultID+"|"+result.LastNode.FileName]
				resultPath := findResultPath(result.PathID, methodLinesByPath)
				if resultPath == nil {
					log.Info().Msgf("Result path not found for ID: %s, on file name: %s and pathId %s",
						result.ResultID, result.FirstNode.FileName, result.PathID)
					continue
				}
				methodLines := resultPath.MethodLines
				similarityCalculationJobs <- SimilarityCalculationJob{
					result.ResultID, result.PathID,
					firstSourceFile.LocalName, result.FirstNode.Name, result.FirstNode.Line, result.FirstNode.Column, methodLines[0],
					lastSourceFile.LocalName, result.LastNode.Name, result.LastNode.Line, result.LastNode.Column, methodLines[len(methodLines)-1],
					astQueryID, e.simIDVersion,
				}
			}
			close(similarityCalculationJobs)
		}()

		// consume calculation jobs
		var resultsMutex sync.Mutex
		similarityCalculationResults := make([]SimilarityCalculationResult, 0, len(query.Results))
		var wg sync.WaitGroup
		for consumerID := 1; consumerID <= worker.GetNumCPU(); consumerID++ {
			wg.Add(1)
			// Worker goroutine
			go func() {
				defer wg.Done()
				for job := range similarityCalculationJobs {
					similarityID, similarityIDErr := e.similarityIDProvider.Calculate(
						job.Filename1, job.Name1, job.Line1, job.Column1, job.MethodLine1,
						job.Filename2, job.Name2, job.Line2, job.Column2, job.MethodLine2,
						job.QueryID,
						job.SimIDVersion,
					)
					result := SimilarityCalculationResult{
						ResultID:     job.ResultID,
						PathID:       job.PathID,
						SimilarityID: similarityID,
						Err:          similarityIDErr,
					}
					resultsMutex.Lock()
					similarityCalculationResults = append(similarityCalculationResults, result)
					resultsMutex.Unlock()
				}
			}()
		}

		// Wait for all workers to finish
		wg.Wait()

		// Sort results by ResultID and PathID
		sort.Slice(similarityCalculationResults, func(i, j int) bool {
			if similarityCalculationResults[i].ResultID == similarityCalculationResults[j].ResultID {
				return similarityCalculationResults[i].PathID < similarityCalculationResults[j].PathID
			}
			return similarityCalculationResults[i].ResultID < similarityCalculationResults[j].ResultID
		})
		// build lookup maps for O(1) access during result handling
		recordResultByID := make(map[string]*RecordResult)
		recordPathByKey := make(map[string]*RecordPath)
		origSimByKey := make(map[string]string)
		for _, orig := range query.Results {
			origSimByKey[orig.ResultID+"|"+orig.PathID] = orig.SimilarityID
		}

		// handle calculation results
		for _, r := range similarityCalculationResults {
			if r.Err != nil {
				return nil, errors.Wrap(r.Err, "failed calculating similarity id")
			}

			recordResult, exists := recordResultByID[r.ResultID]
			if !exists {
				recordResult = &RecordResult{ResultID: r.ResultID}
				output.Queries[queryIdx].Results = append(output.Queries[queryIdx].Results, recordResult)
				recordResultByID[r.ResultID] = recordResult
			}

			pathKey := r.ResultID + "|" + r.PathID
			if _, exists := recordPathByKey[pathKey]; !exists {
				recordPath := &RecordPath{
					PathID:           r.PathID,
					SimilarityID:     r.SimilarityID,
					ResultID:         r.ResultID,
					SASTSimilarityID: origSimByKey[pathKey],
				}
				recordResult.Paths = append(recordResult.Paths, recordPath)
				recordPathByKey[pathKey] = recordPath
			}
		}

		// Sort query.Results by ResultID and PathID
		sort.Slice(query.Results, func(i, j int) bool {
			if query.Results[i].ResultID == query.Results[j].ResultID {
				return query.Results[i].PathID < query.Results[j].PathID
			}
			return query.Results[i].ResultID < query.Results[j].ResultID
		})
	}

	return output, nil
}

func findSourceFile(resultID, remoteName string, sourceFiles []interfaces.SourceFile) *interfaces.SourceFile {
	for _, v := range sourceFiles {
		if v.RemoteName == remoteName && v.ResultID == resultID {
			return &v
		}
	}
	return nil
}

func findResultPath(pathID string, methodLines []*interfaces.ResultPath) *interfaces.ResultPath {
	for _, v := range methodLines {
		if v != nil && strings.TrimSpace(v.PathID) == pathID {
			return v
		}
	}
	return nil
}

func GetQueriesFromReport(reportReader *report.CxXMLResults) []*Query {
	var output []*Query
	for i := 0; i < len(reportReader.Queries); i++ {
		q := reportReader.Queries[i]
		query := &Query{
			QueryID:  q.ID,
			Name:     q.Name,
			Language: q.Language,
			Group:    q.Group,
		}
		for j := 0; j < len(q.Results); j++ {
			r := q.Results[j]
			// only triaged results will have metadata records generated
			if r.State == "0" {
				continue
			}
			for k := 0; k < len(r.Paths); k++ {
				p := r.Paths[k]
				firstNode := p.PathNodes[0]
				lastNode := p.PathNodes[len(p.PathNodes)-1]
				query.Results = append(query.Results, &Result{
					ResultID:     p.ResultID,
					PathID:       p.PathID,
					SimilarityID: p.SimilarityID,
					FirstNode: Node{
						FileName: firstNode.FileName,
						Name:     firstNode.Name,
						Line:     firstNode.Line,
						Column:   firstNode.Column,
					},
					LastNode: Node{
						FileName: lastNode.FileName,
						Name:     lastNode.Name,
						Line:     lastNode.Line,
						Column:   lastNode.Column,
					},
				})
			}
		}
		if len(query.Results) > 0 {
			output = append(output, query)
		}
	}
	return output
}
