package internal

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/app/resultsmapping"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/hashicorp/go-cleanhttp"

	"github.com/checkmarxDev/ast-sast-export/internal/persistence/installation"

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

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/checkmarxDev/ast-sast-export/internal/app/common"
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
	Record    *metadata.Record
}

//nolint:gocyclo,funlen
func RunExport(args *Args) error {
	consumerCount := worker.GetNumCPU()

	log.Debug().
		Str("url", args.URL).
		Str("export", fmt.Sprintf("%v", args.Export)).
		Str("queryMapping", args.QueryMappingFile).
		Str("queryRenaming", args.QueryRenamingFile).
		Int("projectsActiveSince", args.ProjectsActiveSince).
		Str("projectId", args.ProjectsIDs).
		Str("projectTeam", args.TeamName).
		Bool("nestedTeams", args.NestedTeams).
		Bool("debug", args.Debug).
		Int("consumers", consumerCount).
		Msg("starting export")

	err := common.LoadRename(args.QueryRenamingFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize query rename mapping")
	}

	retryHTTPClient := getRetryHTTPClient()
	// create api client
	client, clientErr := rest.NewSASTClient(args.URL, retryHTTPClient)
	if clientErr != nil {
		return errors.Wrap(clientErr, "could not create REST client")
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

	// Fetch custom extensions if provided via cli
	if err := fetchCustomExtensions(args, &exportValues); err != nil {
		return errors.Wrap(err, "failed to fetch custom extensions")
	}

	soapClient := soap.NewClient(args.URL, client.Token, retryHTTPClient)
	sourceRepo := sourcefile.NewRepo(soapClient)
	methodLineRepo := methodline.NewRepo(soapClient)
	queriesRepo := queries.NewRepo(soapClient)
	presetRepo := presetrepo.NewRepo(soapClient)
	installationRepo := installation.NewRepo(soapClient)

	fetchInstallationErr := fetchInstallationData(client, installationRepo, &exportValues)
	if fetchInstallationErr != nil {
		return errors.Wrap(fetchInstallationErr, "could not fetch installation data")
	}

	astQueryMappingProvider, astQueryMappingProviderErr := querymapping.NewProvider(args.QueryMappingFile, retryHTTPClient)
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

	metadataSource := metadata.NewMetadataFactory(
		astQueryProvider,
		similarityIDCalculator,
		sourceRepo,
		methodLineRepo,
		metadataTempDir,
		args.SimIDVersion,
		args.ExcludeFile,
		args.CustomExtensions,
	)

	addErr := addCustomQueryIDs(astQueryProvider, astQueryMappingProvider)
	if addErr != nil {
		return errors.Wrap(addErr, "could not add custom query ids to mapping")
	}

	addFileErr := addQueryMappingFile(astQueryMappingProvider, &exportValues)
	if addFileErr != nil {
		return errors.Wrap(addFileErr, "could not add query mapping file")
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

//nolint:gocyclo
func fetchSelectedData(client rest.Client, exporter export2.Exporter, args *Args, retryAttempts int,
	retryMinSleep, retryMaxSleep time.Duration, metadataProvider metadata.Provider,
	astQueryProvider interfaces.ASTQueryProvider, presetProvider interfaces.PresetProvider,
) error {
	var errProjects error
	var projects []*rest.Project
	options := sliceutils.ConvertStringToInterface(args.Export)
	if sliceutils.Contains(export2.ProjectsOption, options) {
		projects, errProjects = fetchProjectsData(client, exporter, args.ProjectsActiveSince, args.TeamName, args.ProjectsIDs,
			args.IsDefaultProjectActiveSince)
		if errProjects != nil {
			return errProjects
		}
	}
	for _, exportOption := range export2.GetOptions() {
		if sliceutils.Contains(exportOption, options) {
			switch exportOption {
			case export2.UsersOption:
				if err := fetchUsersData(client, exporter, args); err != nil {
					return err
				}
			case export2.TeamsOption:
				if err := fetchTeamsData(client, exporter, args); err != nil {
					return err
				}
			case export2.QueriesOption:
				if err := fetchQueriesData(astQueryProvider, exporter); err != nil {
					return err
				}
			case export2.PresetsOption:
				if err := fetchPresetsData(client, presetProvider, exporter, projects, args.ProjectsIDs); err != nil {
					return err
				}
			case export2.FiltersOption:
				if err := fetchProjectExcludeSettings(client, exporter, projects, args.ProjectsIDs); err != nil {
					return err
				}
			case export2.EngineConfigurationsOption:
				if err := getProjectConfigurations(client, projects, exporter, args.ProjectsIDs); err != nil {
					return err
				}
			case export2.ResultsOption:
				if err := fetchResultsData(client, astQueryProvider, exporter, args.ProjectsActiveSince, retryAttempts, retryMinSleep,
					retryMaxSleep, metadataProvider, args.TeamName, args.ProjectsIDs, args); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func fetchUsersData(client rest.Client, exporter export2.Exporter, args *Args) error {
	log.Info().Msg("collecting users")
	users, usersErr := client.GetUsers()
	if usersErr != nil {
		return errors.Wrap(usersErr, "failed getting users")
	}
	teams, teamsErr := client.GetTeams()
	if teamsErr != nil {
		return errors.Wrap(teamsErr, "failed getting teams")
	}
	usersDataSource := export2.NewJSONDataSource(
		export2.TransformUsers(users, teams, export2.TransformOptions{NestedTeams: args.NestedTeams}),
	)
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

func fetchTeamsData(client rest.Client, exporter export2.Exporter, args *Args) error {
	log.Info().Msg("collecting teams")
	teams, teamsErr := client.GetTeams()
	if teamsErr != nil {
		return errors.Wrap(teamsErr, "failed getting teams")
	}
	transformOptions := export2.TransformOptions{NestedTeams: args.NestedTeams}
	teamsDataSource := export2.NewJSONDataSource(export2.TransformTeams(teams, transformOptions))
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
	samlTeamMappingsDataSource := export2.NewJSONDataSource(export2.TransformSamlTeamMappings(samlTeamMappings, transformOptions))
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
	teamName, projectsIDs string, isDefaultProjectActiveSince bool) ([]*rest.Project, error) {
	log.Info().Msg("collecting projects")
	projects := []*rest.Project{}
	projectOffset := 0
	projectLimit := resultsPageLimit
	fromDate := getDateFrom(resultsProjectActiveSince, isDefaultProjectActiveSince, projectsIDs)
	for {
		log.Debug().
			Str("fromDate", fromDate).
			Int("offset", projectOffset).
			Int("limit", projectLimit).
			Msg("fetching projects with custom fields")
		log.Info().Msg("searching for projects...")

		projectsItems, projectsErr := client.GetProjects(fromDate, teamName, projectsIDs, projectOffset, projectLimit)
		if projectsErr != nil {
			return nil, errors.Wrap(projectsErr, "failed getting projects")
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
		return nil, err
	}
	return projects, nil
}

func fetchInstallationData(restClient rest.Client, soapClient interfaces.InstallationProvider, exporter export2.Exporter) error {
	log.Info().Msg("collecting installation details")
	installationResp, errInstallations := soapClient.GetInstallationSettings()
	if errInstallations != nil {
		return errors.Wrap(errInstallations, "error with getting installation details")
	}
	installations := export2.TransformXMLInstallationMappings(installationResp)

	if !export2.ContainsEngine(export2.InstallationEngineServiceName, installations) {
		engineServersResp, errEngineServers := restClient.GetEngineServers()
		if errEngineServers != nil {
			return errors.Wrap(errEngineServers, "error with getting engine servers details")
		}
		engineServerInstallation := export2.TransformEngineServers(engineServersResp)
		installations = append(installations, engineServerInstallation...)
	}

	installationMappingsDataSource := export2.NewJSONDataSource(installations)
	return exporter.AddFileWithDataSource(export2.InstallationFileName, installationMappingsDataSource)
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

	log.Info().Msg("collecting custom states")
	customStateResp, err := client.GetCustomStatesList()
	if err != nil {
		return errors.Wrap(err, "error with getting custom states list")
	}
	customStateData, marshalErr := xml.MarshalIndent(customStateResp, "  ", "    ")
	if marshalErr != nil {
		return errors.Wrap(marshalErr, "marshal error with getting custom states list")
	}

	if errExp := exporter.AddFile(export2.CustomStatesFileName, customStateData); errExp != nil {
		return errors.Wrap(errExp, "error with exporting custom states list to file")
	}

	return nil
}

func fetchPresetsData(
	client rest.Client,
	soapClient interfaces.PresetProvider,
	exporter export2.Exporter,
	projects []*rest.Project, projectsIDs string) error {
	log.Info().Msg("collecting presets")
	consumerCount := worker.GetNumCPU()
	presetJobs := make(chan PresetJob)

	presetList, listErr := client.GetPresets()
	if listErr != nil {
		return errors.Wrap(listErr, "error with getting preset list")
	}
	if projectsIDs != "" {
		presetList = filterPresetByProjectList(presetList, projects)
		log.Info().Msgf("%d associated presets found", len(presetList))
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
	collectedPresets := []int{}
	for i := 0; i < presetCount; i++ {
		presetOutput := <-presetConsumeOutputs
		if presetOutput.Err == nil {
			collectedPresets = append(collectedPresets, presetOutput.PresetID)
		} else {
			presetConsumeErrorCount++
			log.Warn().
				Int("presetID", presetOutput.PresetID).
				Msgf("failed collecting preset %d/%d", i+1, presetCount)
		}
	}
	log.Info().
		Int("totalCollected", len(collectedPresets)).
		Int("totalFailed", presetConsumeErrorCount).
		Msg("Preset collection summary")

	if presetConsumeErrorCount > 0 {
		log.Warn().Msgf("failed collecting %d/%d presets", presetConsumeErrorCount, presetCount)
	}

	return nil
}

//nolint:gocyclo,funlen
func fetchResultsData(client rest.Client, astQueryProvider interfaces.ASTQueryProvider, exporter export2.Exporter,
	resultsProjectActiveSince int, retryAttempts int, retryMinSleep, retryMaxSleep time.Duration,
	metadataProvider metadata.Provider, teamName, projectsIDs string, args *Args,
) error {
	consumerCount := worker.GetNumCPU()
	reportJobs := make(chan ReportJob)

	// Fetch the state mapping once, before starting the report consumers
	stateMapping, err := astQueryProvider.GetStateMapping()
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch state mapping for all reports")
		// Default to an empty mapping to avoid nil pointer issues and allow processing to continue
		stateMapping = make(map[string]string)
	}

	fromDate := getDateFrom(resultsProjectActiveSince, args.IsDefaultProjectActiveSince, projectsIDs)
	triagedScans, triagedScanErr := getTriagedScans(client, fromDate, teamName, projectsIDs)
	if triagedScanErr != nil {
		return triagedScanErr
	}

	log.Debug().
		Int("count", len(triagedScans)).
		Str("scans", fmt.Sprintf("%v", triagedScans)).
		Msg("last scans by project")

	// create and fetch report for each scan
	go produceReports(triagedScans, reportJobs)

	reportCount := len(triagedScans)
	reportConsumeOutputs := make(chan ReportConsumeOutput, reportCount)

	// Define the report consumer function as a closure
	consumeReportForWorker := func(currentWorkerID int) {
		// This closure captures:
		// client, exporter, args, metadataProvider (from fetchResultsData params)
		// retryAttempts, retryMinSleep, retryMaxSleep (from fetchResultsData params)
		// reportJobs, reportConsumeOutputs (channels defined in fetchResultsData)
		// stateMapping (defined in fetchResultsData)

		for reportJob := range reportJobs {
			l := log.With().
				Int("ProjectID", reportJob.ProjectID).
				Int("ScanID", reportJob.ScanID).
				Int("worker", currentWorkerID).
				Logger()

			// Create scan report
			var reportData []byte
			var reportCreateErr error
			retry := rest.Retry{ // This retry struct was locally defined in original consumeReports
				Attempts: 10, // This is the attempts for client.CreateScanReport's internal retry mechanism
				MinSleep: 1 * time.Second,
				MaxSleep: 5 * time.Minute,
			}
			// The loop for retrying CreateScanReport uses retryAttempts from fetchResultsData
			for i := 1; i <= retryAttempts; i++ {
				reportData, reportCreateErr = client.CreateScanReport(reportJob.ScanID, reportJob.ReportType, retry)
				if reportCreateErr != nil {
					l.Debug().Err(reportCreateErr).
						Int("attempt", i).
						Msg("failed creating scan report")
					time.Sleep(retryablehttp.DefaultBackoff(retryMinSleep, retryMaxSleep, i, nil))
				} else {
					break
				}
			}
			if len(reportData) == 0 {
				// scanReportCreateAttempts is a global constant used for the error message context
				l.Debug().Err(reportCreateErr).Msgf("failed creating scan report after %d attempts", scanReportCreateAttempts)
				reportConsumeOutputs <- ReportConsumeOutput{Err: reportCreateErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
				continue
			}

			// Generate metadata json
			var reportReader report.CxXMLResults
			unmarshalErr := xml.Unmarshal(reportData, &reportReader)
			if unmarshalErr != nil {
				l.Error().Err(unmarshalErr).Msg("Failed to unmarshal XML report data")
				reportConsumeOutputs <- ReportConsumeOutput{Err: unmarshalErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
				continue
			}

			// Prepare data for transformation and export
			dataToTransform := reportData // Default to original data
			if len(stateMapping) > 0 {
				statesToUpdateCount := 0
				statesUpdated := false
				for i := range reportReader.Queries {
					query := &reportReader.Queries[i]
					for j := range query.Results {
						result := &query.Results[j]
						if _, exists := stateMapping[result.State]; exists {
							statesToUpdateCount++
						}
					}
				}

				if statesToUpdateCount > 0 {
					l.Info().Msgf("Found %d results with states that will be updated based on the mapping", statesToUpdateCount)
					for i := range reportReader.Queries {
						query := &reportReader.Queries[i]
						for j := range query.Results {
							result := &query.Results[j]
							if newStateID, exists := stateMapping[result.State]; exists {
								l.Info().Msgf("Found result with state='%s' (NodeId: %s, FileName: %s, Line: %s), updating to state='%s'",
									result.State, result.NodeID, result.FileName, result.Line, newStateID)
								reportReader.Queries[i].Results[j].State = newStateID
								statesUpdated = true
							}
						}
					}

					if statesUpdated {
						var buf bytes.Buffer
						buf.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
`))
						encoder := xml.NewEncoder(&buf)
						encoder.Indent("", "  ")
						if errEnc := encoder.Encode(&reportReader); errEnc != nil {
							l.Debug().Err(errEnc).Msg("failed to encode modified report data into buffer")
							reportConsumeOutputs <- ReportConsumeOutput{Err: errEnc, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
							continue
						}
						dataToTransform = buf.Bytes()
					}
				}
			}

			metadataQueries := metadata.GetQueriesFromReport(&reportReader)
			metadataRecord, metadataRecordErr := metadataProvider.GetMetadataRecord(reportReader.ScanID, metadataQueries)
			if metadataRecordErr != nil {
				l.Debug().Err(metadataRecordErr).Msg("failed creating metadata")
				reportConsumeOutputs <- ReportConsumeOutput{Err: metadataRecordErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
				continue
			}
			metadataRecordJSON, metadataRecordJSONErr := json.Marshal(metadataRecord)
			if metadataRecordJSONErr != nil {
				l.Debug().Err(metadataRecordJSONErr).Msg("failed marshaling metadata")
				reportConsumeOutputs <- ReportConsumeOutput{Err: metadataRecordJSONErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
				continue
			}
			exportMetadataErr := exporter.AddFile(fmt.Sprintf(scansMetadataFileName, reportJob.ProjectID), metadataRecordJSON)
			if exportMetadataErr != nil {
				l.Debug().Err(exportMetadataErr).Msg("failed saving metadata")
				reportConsumeOutputs <- ReportConsumeOutput{Err: exportMetadataErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
				continue
			}

			transformedReportData, transformErr := export2.TransformScanReport(
				dataToTransform, export2.TransformOptions{NestedTeams: args.NestedTeams},
			)
			if transformErr != nil {
				l.Debug().Err(transformErr).Msg("failed transforming report data")
				reportConsumeOutputs <- ReportConsumeOutput{Err: transformErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
				continue
			}
			exportErr := exporter.AddFile(fmt.Sprintf(scansFileName, reportJob.ProjectID), transformedReportData)
			if exportErr != nil {
				l.Debug().Err(exportErr).Msg("failed saving result")
				reportConsumeOutputs <- ReportConsumeOutput{Err: exportErr, ProjectID: reportJob.ProjectID,
					ScanID: reportJob.ScanID, Record: metadataRecord}
			} else {
				reportConsumeOutputs <- ReportConsumeOutput{Err: nil, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID, Record: metadataRecord}
			}
		}
	} // End of consumeReportForWorker closure

	for consumerID := 1; consumerID <= consumerCount; consumerID++ {
		go consumeReportForWorker(consumerID) // Call the closure
	}

	metadataRecord := make([]*metadata.Record, 0)
	reportConsumeErrorCount := 0
	for i := 0; i < reportCount; i++ {
		consumeOutput := <-reportConsumeOutputs
		if consumeOutput.Record != nil {
			metadataRecord = append(metadataRecord, consumeOutput.Record)
		}
		reportIndex := i + 1
		if consumeOutput.Err == nil {
			log.Info().
				Int("projectID", consumeOutput.ProjectID).
				Int("scanID", consumeOutput.ScanID).
				Str("progress", fmt.Sprintf("%d/%d", reportIndex, reportCount)).
				Msg("Successfully collected results")
		} else {
			reportConsumeErrorCount++
			log.Warn().
				Int("projectID", consumeOutput.ProjectID).
				Int("scanID", consumeOutput.ScanID).
				Err(consumeOutput.Err).
				Msg("Failed to collect scan results")
		}
	}

	allResultsMappingErr := addAllResultsMappingToFile(metadataRecord, exporter)
	if allResultsMappingErr != nil {
		log.Debug().Err(allResultsMappingErr).Msg("failed saving results mapping")
	}

	if reportConsumeErrorCount > 0 {
		log.Warn().Msgf("failed collecting %d/%d results", reportConsumeErrorCount, reportCount)
	}

	return nil
}

func addAllResultsMappingToFile(metadataRecord []*metadata.Record, exporter export2.Exporter) error {
	metadataResultsCSV := resultsmapping.GenerateCSV(metadataRecord)
	metadataResultsCSVByte := resultsmapping.WriteAllToSanitizedCsv(metadataResultsCSV)
	exportResultsErr := exporter.AddFile(export2.ResultsMappingFileName, metadataResultsCSVByte)
	if exportResultsErr != nil {
		return exportResultsErr
	}
	log.Info().Msg("collected results mapping")
	return nil
}

func getTriagedScans(client rest.Client, fromDate, teamName, projectsIDs string) ([]TriagedScan, error) {
	var output []TriagedScan
	projectOffset := 0
	projectLimit := resultsPageLimit

	log.Info().Msg("searching for results...")
	for {
		log.Debug().
			Str("fromDate", fromDate).
			Int("offset", projectOffset).
			Int("limit", projectLimit).
			Msg("fetching project last scans")

		// fetch current page
		projects, fetchErr := client.GetProjectsWithLastScanID(fromDate, teamName, projectsIDs, projectOffset,
			projectLimit)
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
				log.Info().Msgf("%d triaged results found from projectId=%d scanId=%d", len(*triagedResults), project.ID, project.LastScanID)
			}
		}

		// prepare to fetch next page
		projectOffset += projectLimit
	}
	return output, nil
}

func getProjectConfigurations(client rest.Client, projects []*rest.Project, exporter export2.Exporter, projectIDs string) error {
	var engineConfigs []JoinedConfig
	log.Info().Msg("collecting engine configurations")

	// Fetch engine configuration mappings
	mappingsData, err := client.GetEngineConfigurationMappings()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get engine configuration mappings")
	}

	// Parse the engine configuration mappings
	var engineMappings []EngineConfigMapping
	if err := json.Unmarshal(mappingsData, &engineMappings); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal engine configuration mappings")
	}

	// Fetch engine keys and create a map for quick lookup by name
	engineKeysMap := readEngineConfigurationsKeys(client)

	// Create a map for quick lookup of engine configuration names by ID
	engineConfigMap := make(map[int]string)
	for _, mapping := range engineMappings {
		engineConfigMap[mapping.EngineConfigurationID] = mapping.Name
	}

	// If specific project IDs are provided, filter the projects
	if projectIDs != "" {
		// Split the project IDs string into individual IDs
		ids := strings.Split(projectIDs, ",")
		idMap := make(map[string]bool)
		for _, id := range ids {
			idMap[id] = true
		}

		// Filter projects based on provided IDs
		for _, project := range projects {
			//nolint:dupl,gocritic
			if idMap[strconv.Itoa(project.ID)] {
				configs, err := client.GetEngineConfigurations(project.ID)
				if err != nil {
					log.Error().Err(err).
						Int("projectID", project.ID).
						Msg("Skipping project due to error with getting engine configurations details")
					continue
				}
				var engineConfiguration rest.EngineConfigurations
				if err := json.Unmarshal(configs, &engineConfiguration); err != nil {
					log.Error().Err(err).Msg("Skipping project due to failed unmarshalling of scan settings")
					continue
				}

				engineConfigName := engineConfigMap[engineConfiguration.EngineConfiguration.ID]
				keys := engineKeysMap[engineConfigName]

				engineConfigs = append(engineConfigs, JoinedConfig{
					ProjectID:               engineConfiguration.Project.ID,
					EngineConfigurationID:   engineConfiguration.EngineConfiguration.ID,
					EngineConfigurationName: engineConfigName,
					ConfigurationKeys:       keys,
				})
			}
		}
	} else {
		// If no specific IDs provided, process all projects
		//nolint:dupl
		for _, project := range projects {
			configs, err := client.GetEngineConfigurations(project.ID)
			if err != nil {
				log.Error().Err(err).
					Int("projectID", project.ID).
					Msg("Skipping project due to error with getting engine configurations details")
				continue
			}
			var engineConfiguration rest.EngineConfigurations
			if err := json.Unmarshal(configs, &engineConfiguration); err != nil {
				log.Error().Err(err).Msg("Skipping project due to failed unmarshalling of scan settings")
				continue
			}

			engineConfigName := engineConfigMap[engineConfiguration.EngineConfiguration.ID]
			keys := engineKeysMap[engineConfigName]
			engineConfigs = append(engineConfigs, JoinedConfig{
				ProjectID:               engineConfiguration.Project.ID,
				EngineConfigurationID:   engineConfiguration.EngineConfiguration.ID,
				EngineConfigurationName: engineConfigName,
				ConfigurationKeys:       keys,
			})
		}
	}

	if len(engineConfigs) > 0 {
		configsDataSource := export2.NewJSONDataSource(engineConfigs)
		exportErr := exporter.AddFileWithDataSource(export2.EngineConfigurationPerProjectFileName, configsDataSource)
		if exportErr != nil {
			log.Error().Err(exportErr).Msg("Failed to export engine configurations data")
			return exportErr
		}
	} else {
		log.Info().Msg("No engine configurations to export")
	}

	return nil
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

func filterPresetByProjectList(presets []*rest.PresetShort, projects []*rest.Project) []*rest.PresetShort {
	out := make([]*rest.PresetShort, 0)
	for _, item := range presets {
		if isIncludedPreset(projects, item) {
			out = append(out, item)
		}
	}
	return out
}

func isIncludedPreset(projects []*rest.Project, value *rest.PresetShort) bool {
	for _, project := range projects {
		if project.PresetID == value.ID {
			return true
		}
	}
	return false
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

func getRetryHTTPClient() *retryablehttp.Client {
	return &retryablehttp.Client{
		HTTPClient:   cleanhttp.DefaultPooledClient(),
		Logger:       nil,
		RetryWaitMin: httpRetryWaitMin,
		RetryWaitMax: httpRetryWaitMax,
		RetryMax:     httpRetryMax,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
		RequestLogHook: func(_ retryablehttp.Logger, request *http.Request, i int) {
			log.Debug().
				Str("method", request.Method).
				Str("url", request.URL.String()).
				Int("attempt", i+1).
				Msg("request")
		},
		ResponseLogHook: func(_ retryablehttp.Logger, response *http.Response) {
			log.Debug().
				Str("method", response.Request.Method).
				Str("url", response.Request.URL.String()).
				Int("status", response.StatusCode).
				Msg("response")
		},
	}
}

func addQueryMappingFile(queryMappingProvider interfaces.QueryMappingRepo, exporter export2.Exporter) error {
	mapping := querymapping.MapSource{
		Mappings: queryMappingProvider.GetMapping(),
	}
	return exporter.AddFileWithDataSource(destQueryMappingFile, export2.NewJSONDataSource(mapping))
}

func addCustomQueryIDs(astQueryProvider interfaces.ASTQueryProvider, astQueryMappingProvider interfaces.QueryMappingRepo) error {
	customQueryResp, errResp := astQueryProvider.GetCustomQueriesList()
	if errResp != nil {
		return errResp
	}
	for i := range customQueryResp.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup {
		queryGroup := customQueryResp.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup[i]
		for j := range queryGroup.Queries.CxWSQuery {
			query := queryGroup.Queries.CxWSQuery[j]
			if err := astQueryMappingProvider.AddQueryMapping(queryGroup.LanguageName, query.Name,
				queryGroup.Name, strconv.Itoa(query.QueryID)); err != nil {
				return err
			}
		}
	}

	return nil
}

func getDateFrom(resultsProjectActiveSince int, isDefaultProjectActiveSince bool, projectIDs string) string {
	if isDefaultProjectActiveSince && projectIDs != "" {
		return ""
	}
	return GetDateFromDays(resultsProjectActiveSince, time.Now())
}

func readEngineConfigurationsKeys(client rest.Client) map[string][]EngineKey {
	mapping, err := client.GetConfigurationsKeys()
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch engine config mapping")
		return map[string][]EngineKey{}
	}

	mappingBytes, err := json.Marshal(mapping)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal engine config mapping")
		return map[string][]EngineKey{}
	}

	var engineKeysData EngineKeysData
	if err := json.Unmarshal(mappingBytes, &engineKeysData); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal engine keys")
		return map[string][]EngineKey{}
	}

	engineKeysMap := make(map[string][]EngineKey)

	for _, config := range engineKeysData.EngineConfig.Configurations.Configuration {
		switch keys := config.Keys.(type) {
		case map[string]interface{}:
			if keyArray, ok := keys["Key"].([]interface{}); ok {
				var engineKeys []EngineKey
				for _, key := range keyArray {
					if keyMap, ok := key.(map[string]interface{}); ok {
						name, nameOk := keyMap["Name"].(string)
						value, valueOk := keyMap["Value"].(string)
						if nameOk && valueOk {
							engineKeys = append(engineKeys, EngineKey{
								Name:  name,
								Value: value,
							})
						}
					}
				}
				engineKeysMap[config.Name] = engineKeys
			}
		default:
			log.Warn().Msg("Keys is of an unexpected type")
		}
	}

	return engineKeysMap
}

func fetchCustomExtensions(args *Args, exporter export2.Exporter) error {
	if args.CustomExtensions == "" {
		return nil
	}

	// Split the input string by space to get language, extension and language group
	parts := strings.Split(args.CustomExtensions, " ")
	if len(parts) < 3 || len(parts)%3 != 0 {
		return fmt.Errorf("invalid custom extensions format. Expected: Language Extension LanguageGroup [Language Extension LanguageGroup ...]")
	}

	customExtensions := &CustomExtensionsList{
		CustomExtension: make([]CustomExtension, 0),
	}

	// Process each set of three parts (Language, Extension, LanguageGroup)
	for i := 0; i < len(parts); i += 3 {
		if i+2 >= len(parts) {
			break
		}

		language := parts[i]
		extension := parts[i+1]
		languageGroup := parts[i+2]

		customExtensions.CustomExtension = append(customExtensions.CustomExtension, CustomExtension{
			Language:      language,
			Extension:     extension,
			LanguageGroup: languageGroup,
		})
	}

	// Marshal the custom extensions to XML
	customExtensionsData, err := xml.MarshalIndent(customExtensions, "  ", "    ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal custom extensions")
	}

	// Save the custom extensions to a file
	if err := exporter.AddFile(export2.CustomExtensionsFileName, customExtensionsData); err != nil {
		return errors.Wrap(err, "failed to save custom extensions")
	}

	return nil
}

func fetchProjectExcludeSettings(client rest.Client, exporter export2.Exporter, projects []*rest.Project, projectIDs string) error {
	var allExcludeSettings []*rest.ProjectExcludeSettings

	// If specific project IDs are provided, filter the projects
	if projectIDs != "" {
		// Split the project IDs string into individual IDs
		ids := strings.Split(projectIDs, ",")
		idMap := make(map[string]bool)
		for _, id := range ids {
			idMap[id] = true
		}

		// Filter projects based on provided IDs
		for _, project := range projects {
			//nolint:dupl,gocritic
			if idMap[strconv.Itoa(project.ID)] {
				log.Info().Int("projectID", project.ID).Msg("collecting project exclude settings")
				excludeSettings, err := client.GetProjectExcludeSettings(project.ID)
				if err != nil {
					return errors.Wrapf(err, "failed to get exclude settings for project %d", project.ID)
				}
				allExcludeSettings = append(allExcludeSettings, excludeSettings)
			}
		}
	} else {
		// If no specific IDs provided, process all projects
		for _, project := range projects {
			log.Info().Int("projectID", project.ID).Msg("collecting project exclude settings")
			excludeSettings, err := client.GetProjectExcludeSettings(project.ID)
			if err != nil {
				return errors.Wrapf(err, "failed to get exclude settings for project %d", project.ID)
			}
			allExcludeSettings = append(allExcludeSettings, excludeSettings)
		}
	}

	// Marshal all settings to JSON
	excludeSettingsData, err := json.MarshalIndent(allExcludeSettings, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal exclude settings")
	}

	// Save all settings to a file
	if err := exporter.AddFile(export2.ProjectExcludeSettingsFileName, excludeSettingsData); err != nil {
		return errors.Wrap(err, "failed to save exclude settings")
	}

	return nil
}
