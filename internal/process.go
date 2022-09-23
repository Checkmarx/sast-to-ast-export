package internal

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/app/astquery"
	export2 "github.com/checkmarxDev/ast-sast-export/internal/app/export"
	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/checkmarxDev/ast-sast-export/internal/app/metadata"
	"github.com/checkmarxDev/ast-sast-export/internal/app/permissions"
	"github.com/checkmarxDev/ast-sast-export/internal/app/preset"
	"github.com/checkmarxDev/ast-sast-export/internal/app/querymapping"
	"github.com/checkmarxDev/ast-sast-export/internal/app/report"
	"github.com/checkmarxDev/ast-sast-export/internal/app/worker"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/rest"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/similarity"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	"github.com/checkmarxDev/ast-sast-export/internal/persistence/methodline"
	presetrepo "github.com/checkmarxDev/ast-sast-export/internal/persistence/preset"
	"github.com/checkmarxDev/ast-sast-export/internal/persistence/queries"
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
	httpRetryWaitMin      = 1 * time.Second //nolint:revive
	httpRetryWaitMax      = 30 * time.Second
	httpRetryMax          = 4

	scanReportCreateAttempts = 10
	scanReportCreateMinSleep = 1 * time.Second
	scanReportCreateMaxSleep = 5 * time.Minute

	destQueryMappingFile = "query_mapping.json"
)

type ReportConsumeOutput struct {
	Err       error
	ProjectID int
	ScanID    int
}

//nolint:funlen
func RunExport(args *Args) error {
	consumerCount := worker.GetNumCPU()

	log.Debug().
		Str("url", args.URL).
		Str("export", fmt.Sprintf("%v", args.Export)).
		Str("queryMapping", args.QueryMappingFile).
		Int("projectsActiveSince", args.ProjectsActiveSince).
		Bool("debug", args.Debug).
		Int("consumers", consumerCount).
		Msg("starting export")

	retryHttpClient := getRetryHttpClient()
	// create api client
	client, err := rest.NewSASTClient(args.URL, retryHttpClient)
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
	exportValues, exportCreateErr := export2.CreateExport(args.ProductName, args.RunTime)
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

	soapClient := soap.NewClient(args.URL, client.Token, retryHttpClient)
	sourceRepo := sourcefile.NewRepo(soapClient)
	methodLineRepo := methodline.NewRepo(soapClient)
	queriesRepo := queries.NewRepo(soapClient)
	presetRepo := presetrepo.NewRepo(soapClient)

	astQueryMappingProvider, astQueryMappingProviderErr := querymapping.NewProvider(args.QueryMappingFile, retryHttpClient)
	if astQueryMappingProviderErr != nil {
		return errors.Wrap(astQueryMappingProviderErr, "could not create AST query mapping provider")
	}

	astQueryProvider, astQueryProviderErr := astquery.NewProvider(queriesRepo, astQueryMappingProvider)
	if astQueryProviderErr != nil {
		return errors.Wrap(astQueryProviderErr, "could not create AST query provider")
	}

	presetProvider := preset.NewProvider(presetRepo)

	similarityIDCalculator, similarityIDCalculatorErr := similarity.NewSimilarityIDCalculator()
	if similarityIDCalculatorErr != nil {
		return errors.Wrap(similarityIDCalculatorErr, "could not create similarity id calculator")
	}

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

	metadataSource := metadata.NewMetadataFactory(astQueryProvider, similarityIDCalculator, sourceRepo, methodLineRepo, metadataTempDir)

	copyErr := copyQueryMappingFile(astQueryMappingProvider, &exportValues)
	if copyErr != nil {
		return errors.Wrap(copyErr, "could not copy query mapping file")
	}

	fetchErr := fetchSelectedData(client, &exportValues, args, scanReportCreateAttempts, scanReportCreateMinSleep,
		scanReportCreateMaxSleep, metadataSource, astQueryProvider, presetProvider)
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

	exportFileName, _, exportErr := exportValues.CreateExportPackage(args.ProductName, args.OutputPath)
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
	astQueryProvider interfaces.ASTQueryProvider, presetProvider interfaces.PresetProvider,
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
			case export2.ProjectsOption:
				if err := fetchProjectsData(client, exporter, args.ProjectsActiveSince, args.TeamName,
					args.ProjectsIds); err != nil {
					return err
				}
			case export2.QueriesOption:
				if err := fetchQueriesData(astQueryProvider, exporter); err != nil {
					return err
				}
			case export2.PresetsOption:
				if err := fetchPresetsData(client, presetProvider, exporter); err != nil {
					return err
				}
			case export2.ResultsOption:
				if err := fetchResultsData(client, exporter, args.ProjectsActiveSince, retryAttempts, retryMinSleep,
					retryMaxSleep, metadataProvider, args.TeamName, args.ProjectsIds); err != nil {
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
	teams, teamsErr := client.GetTeams()
	if teamsErr != nil {
		return errors.Wrap(teamsErr, "failed getting teams")
	}
	usersDataSource := export2.NewJSONDataSource(export2.TransformUsers(users, teams))
	if err := exporter.AddFileWithDataSource(export2.UsersFileName, usersDataSource); err != nil {
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
	teamsDataSource := export2.NewJSONDataSource(export2.TransformTeams(teams))
	if err := exporter.AddFileWithDataSource(export2.TeamsFileName, teamsDataSource); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export2.LdapTeamMappingsFileName, client.GetLdapTeamMappings); err != nil {
		return err
	}
	samlTeamMappings, samlTeamMappingsErr := client.GetSamlTeamMappings()
	if samlTeamMappingsErr != nil {
		return errors.Wrap(samlTeamMappingsErr, "failed getting saml team mappings")
	}
	samlTeamMappingsDataSource := export2.NewJSONDataSource(export2.TransformSamlTeamMappings(samlTeamMappings))
	if err := exporter.AddFileWithDataSource(export2.SamlTeamMappingsFileName, samlTeamMappingsDataSource); err != nil {
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

func fetchProjectsData(client rest.Client, exporter export2.Exporter, resultsProjectActiveSince int,
	teamName, projectsIds string) error {
	log.Info().Msg("collecting projects")
	var projects []*rest.Project
	projectOffset := 0
	projectLimit := resultsPageLimit
	fromDate := GetDateFromDays(resultsProjectActiveSince, time.Now())
	for {
		log.Debug().
			Str("fromDate", fromDate).
			Int("offset", projectOffset).
			Int("limit", projectLimit).
			Msg("fetching projects with custom fields")
		log.Info().Msg("searching for projects...")

		projectsItems, projectsErr := client.GetProjects(fromDate, teamName, projectsIds, projectOffset, projectLimit)
		if projectsErr != nil {
			return errors.Wrap(projectsErr, "failed getting projects")
		}
		if len(projectsItems) == 0 {
			break
		}

		projects = append(projects, projectsItems...)

		// prepare to fetch next page
		projectOffset += projectLimit
	}
	if err := exporter.AddFileWithDataSource(export2.ProjectsFileName,
		export2.NewJSONDataSource(projects)); err != nil {
		return err
	}
	return nil
}

func fetchQueriesData(client interfaces.ASTQueryProvider, exporter export2.Exporter) error {
	log.Info().Msg("collecting custom queries")
	queryResp, err := client.GetCustomQueriesList()
	if err != nil {
		return errors.Wrap(err, "error with getting custom queries list")
	}

	queriesData, marshalErr := xml.MarshalIndent(queryResp, "  ", "    ")
	if marshalErr != nil {
		return errors.Wrap(marshalErr, "marshal error with getting custom queries list")
	}

	if errExp := exporter.AddFile(export2.QueriesFileName, queriesData); errExp != nil {
		return errors.Wrap(errExp, "error with exporting custom queries list to file")
	}

	return nil
}

func fetchPresetsData(client rest.Client, soapClient interfaces.PresetProvider, exporter export2.Exporter) error {
	log.Info().Msg("collecting presets")
	consumerCount := worker.GetNumCPU()
	presetJobs := make(chan PresetJob)

	presetList, listErr := client.GetPresets()
	if listErr != nil {
		return errors.Wrap(listErr, "error with getting preset list")
	}
	if err := exporter.CreateDir(export2.PresetsDirName); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(export2.PresetsFileName,
		export2.NewJSONDataSource(presetList)); err != nil {
		return err
	}

	// create and fetch preset by the list
	go producePresets(presetList, presetJobs)

	presetCount := len(presetList)
	presetConsumeOutputs := make(chan PresetConsumeOutput, presetCount)

	for consumerID := 1; consumerID <= consumerCount; consumerID++ {
		go consumePresets(soapClient, exporter, consumerID, presetJobs, presetConsumeOutputs)
	}

	presetConsumeErrorCount := 0
	for i := 0; i < presetCount; i++ {
		presetOutput := <-presetConsumeOutputs
		presetIndex := i + 1
		if presetOutput.Err == nil {
			log.Info().
				Int("presetID", presetOutput.PresetID).
				Msgf("collected preset %d/%d", presetIndex, presetCount)
		} else {
			presetConsumeErrorCount++
			log.Warn().
				Int("presetID", presetOutput.PresetID).
				Msgf("failed collecting preset %d/%d", presetIndex, presetCount)
		}
	}

	if presetConsumeErrorCount > 0 {
		log.Warn().Msgf("failed collecting %d/%d presets", presetConsumeErrorCount, presetCount)
	}

	return nil
}

func fetchResultsData(client rest.Client, exporter export2.Exporter, resultsProjectActiveSince int,
	retryAttempts int, retryMinSleep, retryMaxSleep time.Duration, metadataProvider metadata.MetadataProvider,
	teamName, projectsIds string,
) error {
	consumerCount := worker.GetNumCPU()
	reportJobs := make(chan ReportJob)

	fromDate := GetDateFromDays(resultsProjectActiveSince, time.Now())
	triagedScans, triagedScanErr := getTriagedScans(client, fromDate, teamName, projectsIds)
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

func getTriagedScans(client rest.Client, fromDate, teamName, projectsIds string) ([]TriagedScan, error) {
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
		projects, fetchErr := client.GetProjectsWithLastScanID(fromDate, teamName, projectsIds,
			projectOffset, projectLimit)
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
		transformedReportData, transformErr := export2.TransformScanReport(reportData)
		if transformErr != nil {
			l.Debug().Err(transformErr).Msg("failed transforming report data")
			done <- ReportConsumeOutput{Err: transformErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
			continue
		}
		exportErr := exporter.AddFile(fmt.Sprintf(scansFileName, reportJob.ProjectID), transformedReportData)
		if exportErr != nil {
			l.Debug().Err(exportErr).Msg("failed saving result")
			done <- ReportConsumeOutput{Err: exportErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
		} else {
			done <- ReportConsumeOutput{Err: nil, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
		}
	}
}

func getPresetFileName(fileName string) string {
	return path.Join(export2.PresetsDirName, fileName)
}

func producePresets(presetList []*rest.PresetShort, presetJobs chan<- PresetJob) {
	for _, v := range presetList {
		presetJobs <- PresetJob{
			PresetID: v.ID,
		}
	}
	close(presetJobs)
}

func consumePresets(soapClient interfaces.PresetProvider, exporter export2.Exporter, workerID int,
	presetJobs <-chan PresetJob, done chan<- PresetConsumeOutput) {
	for presetJob := range presetJobs {
		l := log.With().
			Int("PresetID", presetJob.PresetID).
			Int("worker", workerID).
			Logger()
		presetData, presetErr := getPresetData(soapClient, presetJob.PresetID)
		if presetErr != nil {
			l.Debug().Err(presetErr).Msgf("failed creating preset %d", presetJob.PresetID)
			done <- PresetConsumeOutput{Err: presetErr, PresetID: presetJob.PresetID}
			continue
		}

		if errExp := exporter.AddFile(getPresetFileName(fmt.Sprintf("%d.xml", presetJob.PresetID)), presetData); errExp != nil {
			err := errors.Wrapf(errExp, "error with exporting preset %d list to file", presetJob.PresetID)
			done <- PresetConsumeOutput{Err: err, PresetID: presetJob.PresetID}
			continue
		}

		done <- PresetConsumeOutput{Err: nil, PresetID: presetJob.PresetID}
	}
}

func getPresetData(soapClient interfaces.PresetProvider, presetID int) ([]byte, error) {
	presetResponse, err := soapClient.GetPresetDetails(presetID)
	if err != nil {
		return nil, errors.Wrap(err, "error with getting getPresetDetails")
	}

	presetData, marshalErr := xml.MarshalIndent(presetResponse, "  ", "    ")
	if marshalErr != nil {
		return nil, errors.Wrapf(marshalErr, "marshal error with getting preset %d", presetID)
	}
	return presetData, nil
}

func getRetryHttpClient() *retryablehttp.Client {
	return &retryablehttp.Client{
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
	}
}

func copyQueryMappingFile(queryMappingProvider interfaces.QueryMappingRepo, exporter export2.Exporter) error {
	copyErr := exporter.CopyFile(destQueryMappingFile, queryMappingProvider.GetQueryMappingFilePath())
	delErr := queryMappingProvider.Clean()
	if copyErr == nil && delErr == nil {
		return nil
	}

	err := errors.New("Copy query mapping file error")
	if copyErr != nil {
		err = copyErr
	}
	if delErr != nil {
		err = errors.Wrap(delErr, err.Error())
	}

	return err
}
