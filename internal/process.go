package internal

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/app/worker"

	"github.com/checkmarxDev/ast-sast-export/internal/persistence/methodline"

	export2 "github.com/checkmarxDev/ast-sast-export/internal/app/export"
	"github.com/checkmarxDev/ast-sast-export/internal/app/metadata"
	"github.com/checkmarxDev/ast-sast-export/internal/app/permissions"
	"github.com/checkmarxDev/ast-sast-export/internal/app/report"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/rest"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/similarity"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	"github.com/checkmarxDev/ast-sast-export/internal/persistence/astquery"
	"github.com/checkmarxDev/ast-sast-export/internal/persistence/sourcefile"
	"github.com/checkmarxDev/ast-sast-export/pkg/sliceutils"

	"github.com/golang-jwt/jwt"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
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

func RunExport(args *Args) error {
	consumerCount := worker.GetNumCPU()

	log.Debug().
		Str("url", args.URL).
		Str("export", fmt.Sprintf("%v", args.Export)).
		Int("projectsActiveSince", args.ProjectsActiveSince).
		Bool("debug", args.Debug).
		Int("consumers", consumerCount).
		Msg("starting export")

	// create api client
	client, err := rest.NewSASTClient(args.URL, &retryablehttp.Client{
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
		return errors.Wrap(err, "could not create REST client")
	}

	// authenticate
	log.Info().Msg("connecting to SAST")
	if authErr := client.Authenticate(args.Username, args.Password); authErr != nil {
		return errors.Wrap(authErr, "could not authenticate with SAST API")
	}

	// validate permissions
	jwtClaims := jwt.MapClaims{}
	_, _, jwtErr := new(jwt.Parser).ParseUnverified(client.Token.AccessToken, jwtClaims)
	if jwtErr != nil {
		return errors.Wrap(jwtErr, "permissions error - could not parse token")
	}
	permissionsValidateErr := validatePermissions(jwtClaims, args.Export)
	if permissionsValidateErr != nil {
		panic(fmt.Errorf("permissions error - %s", permissionsValidateErr.Error()))
	}

	// collect export data
	log.Info().Msg("collecting data from SAST")
	exportValues, exportCreateErr := export2.CreateExport(args.ProductName)
	if exportCreateErr != nil {
		return errors.Wrap(exportCreateErr, "could not create export package")
	}

	if !args.Debug {
		defer func(exportValues export2.Exporter) {
			cleanErr := exportValues.Clean()
			if cleanErr != nil {
				log.Error().Err(cleanErr).Msg("error cleaning export temporary folder")
			}
		}(&exportValues)
	}

	astQueryIDRepo, astQueryIDRepoErr := astquery.NewRepo(astquery.AllQueries)
	if astQueryIDRepoErr != nil {
		return errors.Wrap(astQueryIDRepoErr, "could not create AST query id repo")
	}

	similarityIDCalculator, similarityIDCalculatorErr := similarity.NewSimilarityIDCalculator()
	if similarityIDCalculatorErr != nil {
		return errors.Wrap(similarityIDCalculatorErr, "could not create similarity id calculator")
	}

	soapClient := soap.NewClient(args.URL, client.Token, &http.Client{})
	sourceRepo := sourcefile.NewRepo(soapClient)
	methodLineRepo := methodline.NewRepo(soapClient)

	metadataTempDir, metadataTempDirErr := os.MkdirTemp("", args.ProductName)
	if metadataTempDirErr != nil {
		return errors.Wrap(metadataTempDirErr, "could not create metadata temporary folder")
	}
	defer func() {
		metadataTempDirRemoveErr := os.RemoveAll(metadataTempDir)
		if metadataTempDirRemoveErr != nil {
			log.Error().Err(metadataTempDirRemoveErr)
		}
	}()

	metadataSource := metadata.NewMetadataFactory(astQueryIDRepo, similarityIDCalculator, sourceRepo, methodLineRepo, metadataTempDir)

	fetchErr := fetchSelectedData(client, &exportValues, args, scanReportCreateAttempts, scanReportCreateMinSleep,
		scanReportCreateMaxSleep, metadataSource)
	if fetchErr != nil {
		return errors.Wrap(fetchErr, "could not fetch selected data")
	}

	// export data to file
	log.Info().Msg("exporting collected data")
	exportFileName, exportErr := exportResultsToFile(args, &exportValues)
	if exportErr != nil {
		log.Error().Err(exportErr).Msg("error exporting collected data")
	}

	log.Info().Msgf("export completed to %s", exportFileName)
	return nil
}

func exportResultsToFile(args *Args, exportValues export2.Exporter) (string, error) {
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

func fetchSelectedData(client rest.Client, exporter export2.Exporter, args *Args, retryAttempts int,
	retryMinSleep, retryMaxSleep time.Duration, metadataProvider metadata.MetadataProvider,
) error {
	options := sliceutils.ConvertStringToInterface(args.Export)
	for _, exportOption := range export2.GetOptions() {
		if sliceutils.Contains(exportOption, options) {
			switch exportOption {
			case export2.UsersOption:
				if err := fetchUsersData(client, exporter); err != nil {
					return err
				}
			case export2.TeamsOption:
				if err := fetchTeamsData(client, exporter); err != nil {
					return err
				}
			case export2.ResultsOption:
				if err := fetchResultsData(client, exporter, args.ProjectsActiveSince, retryAttempts, retryMinSleep,
					retryMaxSleep, metadataProvider); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func fetchUsersData(client rest.Client, exporter export2.Exporter) error {
	log.Info().Msg("collecting users")
	users, usersErr := client.GetUsers()
	if usersErr != nil {
		return errors.Wrap(usersErr, "failed getting users")
	}
	if err := exporter.AddFileWithDataSource(export2.UsersFileName, export2.NewMarshalDataSource(users)); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export2.RolesFileName, client.GetRoles); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export2.LdapRoleMappingsFileName, client.GetLdapRoleMappings); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export2.SamlRoleMappingsFileName, client.GetSamlRoleMappings); err != nil {
		return err
	}
	if _, fileErr := os.Stat(export2.LdapServersFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(export2.LdapServersFileName, client.GetLdapServers); err != nil {
			return err
		}
	}
	if _, fileErr := os.Stat(export2.SamlIdpFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(export2.SamlIdpFileName, client.GetSamlIdentityProviders); err != nil {
			return err
		}
	}
	return nil
}

func fetchTeamsData(client rest.Client, exporter export2.Exporter) error {
	log.Info().Msg("collecting teams")
	teams, teamsErr := client.GetTeams()
	if teamsErr != nil {
		return errors.Wrap(teamsErr, "failed getting teams")
	}
	if err := exporter.AddFileWithDataSource(export2.TeamsFileName, export2.NewMarshalDataSource(teams)); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export2.LdapTeamMappingsFileName, client.GetLdapTeamMappings); err != nil {
		return err
	}
	samlTeamMappings, samlTeamMappingsErr := client.GetSamlTeamMappings()
	if samlTeamMappingsErr != nil {
		return errors.Wrap(samlTeamMappingsErr, "failed getting saml team mappings")
	}
	if err := exporter.AddFileWithDataSource(export2.SamlTeamMappingsFileName, export2.NewMarshalDataSource(samlTeamMappings)); err != nil {
		return err
	}
	if _, fileErr := os.Stat(export2.LdapServersFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(export2.LdapServersFileName, client.GetLdapServers); err != nil {
			return err
		}
	}
	if _, fileErr := os.Stat(export2.SamlIdpFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(export2.SamlIdpFileName, client.GetSamlIdentityProviders); err != nil {
			return err
		}
	}
	return nil
}

func fetchResultsData(client rest.Client, exporter export2.Exporter, resultsProjectActiveSince int,
	retryAttempts int, retryMinSleep, retryMaxSleep time.Duration, metadataProvider metadata.MetadataProvider,
) error {
	consumerCount := worker.GetNumCPU()
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

func getTriagedScans(client rest.Client, fromDate string) ([]TriagedScan, error) {
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
			ReportType: rest.ScanReportTypeXML,
		}
	}
	close(reportJobs)
}

func consumeReports(client rest.Client, exporter export2.Exporter, workerID int,
	reportJobs <-chan ReportJob, done chan<- ReportConsumeOutput, maxAttempts int,
	attemptMinSleep, attemptMaxSleep time.Duration, metadataProvider metadata.MetadataProvider,
) {
	for reportJob := range reportJobs {
		l := log.With().
			Int("ProjectID", reportJob.ProjectID).
			Int("ScanID", reportJob.ScanID).
			Int("worker", workerID).
			Logger()

		// create scan report
		var reportData []byte
		var reportCreateErr error
		retry := rest.Retry{
			Attempts: 10,
			MinSleep: 1 * time.Second,
			MaxSleep: 5 * time.Minute,
		}
		for i := 1; i <= maxAttempts; i++ {
			reportData, reportCreateErr = client.CreateScanReport(reportJob.ScanID, reportJob.ReportType, retry)
			if reportCreateErr != nil {
				l.Debug().Err(reportCreateErr).
					Int("attempt", i).
					Msg("failed creating scan report")
				time.Sleep(retryablehttp.DefaultBackoff(attemptMinSleep, attemptMaxSleep, i, nil))
			} else {
				break
			}
		}
		if len(reportData) == 0 {
			l.Debug().Err(reportCreateErr).Msgf("failed creating scan report after %d attempts", scanReportCreateAttempts)
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
		metadataQueries := metadata.GetQueriesFromReport(&reportReader)
		metadataRecord, metadataRecordErr := metadataProvider.GetMetadataRecord(reportReader.ScanID, metadataQueries)
		if metadataRecordErr != nil {
			l.Debug().Err(metadataRecordErr).Msg("failed creating metadata")
			done <- ReportConsumeOutput{Err: metadataRecordErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
			continue
		} else {
			metadataRecordJSON, metadataRecordJSONErr := json.Marshal(metadataRecord)
			if metadataRecordJSONErr != nil {
				l.Debug().Err(metadataRecordJSONErr).Msg("failed marshaling metadata")
				done <- ReportConsumeOutput{Err: metadataRecordJSONErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
				continue
			}
			exportMetadataErr := exporter.AddFile(fmt.Sprintf(scansMetadataFileName, reportJob.ProjectID), metadataRecordJSON)
			if exportMetadataErr != nil {
				l.Debug().Err(metadataRecordJSONErr).Msg("failed saving metadata")
				done <- ReportConsumeOutput{Err: metadataRecordJSONErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
				continue
			}
		}
		// export report
		exportErr := exporter.AddFile(fmt.Sprintf(scansFileName, reportJob.ProjectID), reportData)
		if exportErr != nil {
			l.Debug().Err(exportErr).Msg("failed saving result")
			done <- ReportConsumeOutput{Err: exportErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
		} else {
			done <- ReportConsumeOutput{Err: nil, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
		}
	}
}
