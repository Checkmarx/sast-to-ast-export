package internal

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/export"
	"github.com/checkmarxDev/ast-sast-export/internal/permissions"
	"github.com/checkmarxDev/ast-sast-export/internal/sast"
	"github.com/checkmarxDev/ast-sast-export/internal/sast/report"
	"github.com/checkmarxDev/ast-sast-export/internal/sliceutils"
	"github.com/checkmarxDev/ast-sast-export/internal/utils"
	"github.com/golang-jwt/jwt"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

const (
	scansFileName         = "%d.xml"
	scansMetadataFileName = "%d.json"
	resultsPageLimit      = 10000
	httpRetryWaitMin      = 1 * time.Second
	httpRetryWaitMax      = 30 * time.Second
	httpRetryMax          = 4

	scanReportCreateAttempts = 10
	scanReportCreateMinSleep = 1 * time.Second
	scanReportCreateMaxSleep = 5 * time.Minute
)

type ReportConsumeOutput struct {
	Err       error
	ProjectID int
	ScanID    int
}

func RunExport(args *Args) {
	consumerCount := GetNumCPU()

	log.Debug().
		Str("url", args.URL).
		Str("export", fmt.Sprintf("%v", args.Export)).
		Int("projectsActiveSince", args.ProjectsActiveSince).
		Bool("debug", args.Debug).
		Int("consumers", consumerCount).
		Msg("starting export")

	/*	// create db connection
		db, dbErr := database.Connect(args.DBConnectionString)
		if dbErr != nil {
			log.Error().Err(dbErr)
			panic(dbErr)
		}



		reportEnricher := report.NewReport(
			report.NewSource(
				store.NewComponentConfigurationStore(db),
				store.NewTaskScans(db),
			),
			store.NewNodeResults(db),
			similarityCalculator,
		)*/
	/*
		similarityCalculator, similarityCalculatorErr := sast.NewSimilarityIDCalculator()
		if similarityCalculatorErr != nil {
			log.Error().Err(similarityCalculatorErr)
			panic(similarityCalculatorErr)
		}
	*/

	// create api client
	client, err := sast.NewSASTClient(args.URL, &retryablehttp.Client{
		HTTPClient:   cleanhttp.DefaultPooledClient(),
		Logger:       nil,
		RetryWaitMin: httpRetryWaitMin,
		RetryWaitMax: httpRetryWaitMax,
		RetryMax:     httpRetryMax,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
		RequestLogHook: func(logger retryablehttp.Logger, request *http.Request, i int) {
			log.Debug().
				Str("method", request.Method).
				Str("url", request.URL.String()).
				Int("attempt", i+1).
				Msg("request")
		},
		ResponseLogHook: func(logger retryablehttp.Logger, response *http.Response) {
			log.Debug().
				Str("method", response.Request.Method).
				Str("url", response.Request.URL.String()).
				Int("status", response.StatusCode).
				Msg("response")
		},
	})
	if err != nil {
		log.Error().Err(err)
		panic(err)
	}

	// authenticate
	log.Info().Msg("connecting to SAST")
	if authErr := client.Authenticate(args.Username, args.Password); authErr != nil {
		log.Error().Err(authErr)
		panic(authErr)
	}

	// validate permissions
	jwtClaims := jwt.MapClaims{}
	_, _, jwtErr := new(jwt.Parser).ParseUnverified(client.Token.AccessToken, jwtClaims)
	if jwtErr != nil {
		log.Error().Err(jwtErr)
		panic(fmt.Errorf("permissions error - could not parse token"))
	}
	permissionsValidateErr := validatePermissions(jwtClaims, args.Export)
	if permissionsValidateErr != nil {
		panic(fmt.Errorf("permissions error - %s", permissionsValidateErr.Error()))
	}

	// collect export data
	log.Info().Msg("collecting data from SAST")
	exportValues, exportCreateErr := export.CreateExport(args.ProductName)
	if exportCreateErr != nil {
		log.Error().Err(exportCreateErr)
		panic(exportCreateErr)
	}

	if !args.Debug {
		defer func(exportValues export.Exporter) {
			cleanErr := exportValues.Clean()
			if cleanErr != nil {
				log.Error().Err(cleanErr).Msg("error cleaning export temporary folder")
			}
		}(&exportValues)
	}

	// FIXME
	fetchErr := fetchSelectedData(client, &exportValues, args, scanReportCreateAttempts, scanReportCreateMinSleep,
		scanReportCreateMaxSleep, export.NewMetadataSource(nil, nil, nil, ""))
	if fetchErr != nil {
		log.Error().Err(fetchErr)
		panic(fmt.Errorf("fetch error - %s", fetchErr.Error()))
	}

	// export data to file
	log.Info().Msg("exporting collected data")
	exportFileName, exportErr := exportResultsToFile(args, &exportValues)
	if exportErr != nil {
		log.Error().Err(exportErr).Msg("error exporting collected data")
	}

	log.Info().Msgf("export completed to %s", exportFileName)
}

func exportResultsToFile(args *Args, exportValues export.Exporter) (string, error) {
	// create export package
	tmpDir := exportValues.GetTmpDir()
	if args.Debug {
		if err := OpenPathInExplorer(tmpDir); err != nil {
			log.Debug().Err(err)
		}
		return tmpDir, nil
	}

	exportFileName, exportErr := exportValues.CreateExportPackage(args.ProductName, args.OutputPath)
	if exportErr != nil {
		return exportFileName, exportErr
	}
	return exportFileName, exportErr
}

func validatePermissions(jwtClaims jwt.MapClaims, selectedExportOptions []string) error {
	claimKeys := []string{"sast-permissions", "access-control-permissions"}
	available, availableErr := permissions.GetFromJwtClaims(jwtClaims, claimKeys)
	if availableErr != nil {
		return availableErr
	}
	required := permissions.GetFromExportOptions(selectedExportOptions)
	missing := permissions.GetMissing(required, available)
	if len(missing) > 0 {
		for _, permission := range missing {
			description, descriptionErr := permissions.GetDescription(permission)
			if descriptionErr != nil {
				description = permission.(string)
				log.Debug().Err(descriptionErr).Msg("could not get permission description")
			}
			log.Error().Msgf("missing permission %s", description)
		}
		return fmt.Errorf("please add missing permissions to your SAST user")
	}
	return nil
}

func fetchSelectedData(client sast.Client, exporter export.Exporter, args *Args, retryAttempts int,
	retryMinSleep, retryMaxSleep time.Duration, metadataProvider export.MetadataProvider,
) error {
	options := sliceutils.ConvertStringToInterface(args.Export)
	for _, exportOption := range export.GetOptions() {
		if sliceutils.Contains(exportOption, options) {
			switch exportOption {
			case export.UsersOption:
				if err := fetchUsersData(client, exporter); err != nil {
					return err
				}
			case export.TeamsOption:
				if err := fetchTeamsData(client, exporter); err != nil {
					return err
				}
			case export.ResultsOption:
				if err := fetchResultsData(client, exporter, args.ProjectsActiveSince, retryAttempts, retryMinSleep,
					retryMaxSleep, metadataProvider); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func fetchUsersData(client sast.Client, exporter export.Exporter) error {
	log.Info().Msg("collecting users")
	if err := exporter.AddFileWithDataSource(export.UsersFileName, client.GetUsers); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export.RolesFileName, client.GetRoles); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export.LdapRoleMappingsFileName, client.GetLdapRoleMappings); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export.SamlRoleMappingsFileName, client.GetSamlRoleMappings); err != nil {
		return err
	}
	if _, fileErr := os.Stat(export.LdapServersFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(export.LdapServersFileName, client.GetLdapServers); err != nil {
			return err
		}
	}
	if _, fileErr := os.Stat(export.SamlIdpFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(export.SamlIdpFileName, client.GetSamlIdentityProviders); err != nil {
			return err
		}
	}
	return nil
}

func fetchTeamsData(client sast.Client, exporter export.Exporter) error {
	log.Info().Msg("collecting teams")
	if err := exporter.AddFileWithDataSource(export.TeamsFileName, client.GetTeams); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export.LdapTeamMappingsFileName, client.GetLdapTeamMappings); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export.SamlTeamMappingsFileName, client.GetSamlTeamMappings); err != nil {
		return err
	}
	if _, fileErr := os.Stat(export.LdapServersFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(export.LdapServersFileName, client.GetLdapServers); err != nil {
			return err
		}
	}
	if _, fileErr := os.Stat(export.SamlIdpFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(export.SamlIdpFileName, client.GetSamlIdentityProviders); err != nil {
			return err
		}
	}
	return nil
}

func fetchResultsData(client sast.Client, exporter export.Exporter, resultsProjectActiveSince int,
	retryAttempts int, retryMinSleep, retryMaxSleep time.Duration, metadataProvider export.MetadataProvider,
) error {
	consumerCount := GetNumCPU()
	reportJobs := make(chan ReportJob)

	fromDate := GetDateFromDays(resultsProjectActiveSince, time.Now())
	triagedScans, triagedScanErr := getTriagedScans(client, fromDate)
	if triagedScanErr != nil {
		return triagedScanErr
	}

	log.Debug().
		Int("count", len(triagedScans)).
		Str("scans", fmt.Sprintf("%v", triagedScans)).
		Msg("last scans by project")

	log.Info().Msgf("%d results found", len(triagedScans))

	// create and fetch report for each scan
	go produceReports(triagedScans, reportJobs)

	reportCount := len(triagedScans)
	reportConsumeOutputs := make(chan ReportConsumeOutput, reportCount)

	for consumerID := 1; consumerID <= consumerCount; consumerID++ {
		go consumeReports(client, exporter, consumerID, reportJobs, reportConsumeOutputs,
			retryAttempts, retryMinSleep, retryMaxSleep, metadataProvider)
	}

	reportConsumeErrorCount := 0
	for i := 0; i < reportCount; i++ {
		consumeOutput := <-reportConsumeOutputs
		reportIndex := i + 1
		if consumeOutput.Err == nil {
			log.Info().
				Int("projectID", consumeOutput.ProjectID).
				Int("scanID", consumeOutput.ScanID).
				Msgf("collected result %d/%d", reportIndex, reportCount)
		} else {
			reportConsumeErrorCount++
			log.Warn().
				Int("projectID", consumeOutput.ProjectID).
				Int("scanID", consumeOutput.ScanID).
				Msgf("failed collecting result %d/%d", reportIndex, reportCount)
		}
	}

	if reportConsumeErrorCount > 0 {
		log.Warn().Msgf("failed collecting %d/%d results", reportConsumeErrorCount, reportCount)
	}

	return nil
}

func getTriagedScans(client sast.Client, fromDate string) ([]TriagedScan, error) {
	var output []TriagedScan
	projectOffset := 0
	projectLimit := resultsPageLimit

	for {
		log.Debug().
			Str("fromDate", fromDate).
			Int("offset", projectOffset).
			Int("limit", projectLimit).
			Msg("fetching project last scans")
		log.Info().Msg("searching for results...")

		// fetch current page
		projects, fetchErr := client.GetProjectsWithLastScanID(fromDate, projectOffset, projectLimit)
		if fetchErr != nil {
			log.Debug().Err(fetchErr).Msg("failed fetching project last scans")
			return output, fmt.Errorf("error searching for results")
		}
		if len(*projects) == 0 {
			// all pages fetched
			break
		}
		// process current page
		log.Debug().
			Int("count", len(*projects)).
			Msg("processing project last scans")

		for _, project := range *projects {
			// get triaged results
			triagedResults, triagedResultsErr := client.GetTriagedResultsByScanID(project.LastScanID)
			if triagedResultsErr != nil {
				log.Debug().Err(triagedResultsErr).
					Int("projectID", project.ID).
					Int("scanID", project.LastScanID).
					Msg("failed fetching triaged results")
				return output, triagedResultsErr
			}
			if len(*triagedResults) > 0 {
				output = append(output, TriagedScan{project.ID, project.LastScanID})
			}
		}

		// prepare to fetch next page
		projectOffset += projectLimit
	}
	return output, nil
}

func produceReports(triagedScans []TriagedScan, reportJobs chan<- ReportJob) {
	for _, scan := range triagedScans {
		reportJobs <- ReportJob{
			ProjectID:  scan.ProjectID,
			ScanID:     scan.ScanID,
			ReportType: sast.ScanReportTypeXML,
		}
	}
	close(reportJobs)
}

func consumeReports(client sast.Client, exporter export.Exporter, worker int,
	reportJobs <-chan ReportJob, done chan<- ReportConsumeOutput, maxAttempts int,
	attemptMinSleep, attemptMaxSleep time.Duration, metadataProvider export.MetadataProvider,
) {
	var resultMetadata []export.MetadataRecord

	for reportJob := range reportJobs {
		// create scan report
		var reportData []byte
		var reportCreateErr error
		retry := utils.Retry{
			Attempts: 10,
			MinSleep: 1 * time.Second,
			MaxSleep: 5 * time.Minute,
		}
		for i := 1; i <= maxAttempts; i++ {
			reportData, reportCreateErr = client.CreateScanReport(reportJob.ScanID, reportJob.ReportType, retry)
			if reportCreateErr != nil {
				log.Debug().Err(reportCreateErr).
					Int("ProjectID", reportJob.ProjectID).
					Int("ScanID", reportJob.ScanID).
					Int("worker", worker).
					Int("attempt", i).
					Msg("failed creating scan report")
				time.Sleep(retryablehttp.DefaultBackoff(attemptMinSleep, attemptMaxSleep, i, nil))
			} else {
				break
			}
		}
		if len(reportData) == 0 {
			log.Debug().Err(reportCreateErr).
				Int("ProjectID", reportJob.ProjectID).
				Int("ScanID", reportJob.ScanID).
				Int("worker", worker).
				Msgf("failed creating scan report after %d attempts", scanReportCreateAttempts)
			done <- ReportConsumeOutput{Err: reportCreateErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
			continue
		}
		// generate metadata json
		var reportReader report.CxXMLResults
		unmarshalErr := xml.Unmarshal(reportData, &reportReader)
		if unmarshalErr != nil {
			done <- ReportConsumeOutput{Err: unmarshalErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
			continue
		}
		for i := 0; i < len(reportReader.Queries); i++ {
			for j := 0; j < len(reportReader.Queries[i].Results); j++ {
				for k := 0; k < len(reportReader.Queries[i].Results[j].Paths); k++ {
					query := reportReader.Queries[i]
					path := reportReader.Queries[i].Results[j].Paths[k]
					firstPathNode := path.PathNodes[0]
					lastPathNode := path.PathNodes[len(path.PathNodes)-1]
					metaQuery := &export.MetadataQuery{
						QueryID:  query.ID,
						Name:     query.Name,
						Language: query.Language,
						Group:    query.Group,
					}
					metaResult := &export.MetadataResult{
						ResultID: path.ResultID,
						PathID:   path.PathID,
						FirstNode: export.MetadataNode{
							FileName: firstPathNode.FileName,
							Name:     firstPathNode.Name,
							Line:     firstPathNode.Line,
							Column:   firstPathNode.Column,
						},
						LastNode: export.MetadataNode{
							FileName: lastPathNode.FileName,
							Name:     lastPathNode.Name,
							Line:     lastPathNode.Line,
							Column:   lastPathNode.Column,
						},
					}
					metadata, metadataErr := metadataProvider.GetMetadataForQueryAndResult(reportReader.ScanID, metaQuery, metaResult)
					if metadataErr != nil {
						panic(metadataErr) //FIXME
					}
					resultMetadata = append(resultMetadata, *metadata)
				}
			}
		}
		resultMetadataJSON, resultMetadataJSONErr := json.Marshal(resultMetadata)
		if resultMetadataJSONErr != nil {
			panic(resultMetadataJSONErr) // FIXME
		}
		exportMetadataErr := exporter.AddFile(fmt.Sprintf(scansMetadataFileName, reportJob.ProjectID), resultMetadataJSON)
		if exportMetadataErr != nil {
			panic(exportMetadataErr) // FIXME
		}
		// export report
		exportErr := exporter.AddFile(fmt.Sprintf(scansFileName, reportJob.ProjectID), reportData)
		if exportErr != nil {
			log.Debug().Err(exportErr).
				Int("ProjectID", reportJob.ProjectID).
				Int("ScanID", reportJob.ScanID).
				Int("worker", worker).
				Msg("failed saving result")
			done <- ReportConsumeOutput{Err: exportErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
		} else {
			done <- ReportConsumeOutput{Err: nil, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
		}
	}
}
