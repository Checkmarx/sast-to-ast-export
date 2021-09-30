package internal

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/export"
	"github.com/checkmarxDev/ast-sast-export/internal/permissions"
	"github.com/checkmarxDev/ast-sast-export/internal/sliceutils"
	"github.com/golang-jwt/jwt"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
)

const (
	scansFileName    = "%d.xml"
	resultsPageLimit = 10000
	httpRetryWaitMin = 1 * time.Second
	httpRetryWaitMax = 30 * time.Second
	httpRetryMax     = 4
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

	// create api client
	client, err := NewSASTClient(args.URL, &retryablehttp.Client{
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
	permissionsValidateErr := validatePermissions(client, args.Export)
	if permissionsValidateErr != nil {
		panic(fmt.Errorf("permissions error - %s", permissionsValidateErr.Error()))
	}

	// collect export data
	log.Info().Msg("collecting data from SAST")
	exportValues, exportCreateErr := CreateExport(args.ProductName)
	if exportCreateErr != nil {
		log.Error().Err(exportCreateErr)
		panic(exportCreateErr)
	}

	if !args.Debug {
		defer func(exportValues *Export) {
			cleanErr := exportValues.Clean()
			if cleanErr != nil {
				log.Error().Err(cleanErr).Msg("error cleaning export temporary folder")
			}
		}(&exportValues)
	}

	fetchErr := fetchSelectedData(client, &exportValues, args)
	if fetchErr != nil {
		log.Error().Err(fetchErr)
		panic(fmt.Errorf("fetch error - %s", fetchErr.Error()))
	}

	// export data to file
	log.Info().Msg("exporting collected data")
	exportFileName, exportErr := ExportResultsToFile(args, &exportValues)
	if exportErr != nil {
		log.Error().Err(exportErr).Msg("error exporting collected data")
	}

	log.Info().Msgf("export completed to %s", *exportFileName)
}

func ExportResultsToFile(args *Args, exportValues *Export) (*string, error) {
	// create export package
	if args.Debug {
		if runtime.GOOS == "windows" {
			cmdErr := exec.Command(`explorer`, exportValues.TmpDir).Run() //nolint:gosec
			// ignore exit status 1, it was being returned even on success
			if cmdErr != nil && cmdErr.Error() != "exit status 1" {
				log.Debug().Err(cmdErr).Msg("could not open temporary folder")
			}
		}
		return &exportValues.TmpDir, nil
	}

	exportFileName, exportErr := exportValues.CreateExportPackage(args.ProductName, args.OutputPath)
	if exportErr != nil {
		return nil, exportErr
	}
	return &exportFileName, exportErr
}

func validatePermissions(client *SASTClient, selectedExportOptions []string) error {
	jwtClaims := jwt.MapClaims{}
	_, _, jwtErr := new(jwt.Parser).ParseUnverified(client.Token.AccessToken, jwtClaims)
	if jwtErr != nil {
		return jwtErr
	}
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

func fetchSelectedData(client *SASTClient, exporter *Export, args *Args) error {
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
				if err := fetchResultsData(client, exporter, args.ProjectsActiveSince); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func fetchUsersData(client *SASTClient, exporter *Export) error {
	log.Info().Msg("collecting users")
	if err := exporter.AddFileWithDataSource(UsersFileName, client.GetUsers); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(RolesFileName, client.GetRoles); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(LdapRoleMappingsFileName, client.GetLdapRoleMappings); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(SamlRoleMappingsFileName, client.GetSamlRoleMappings); err != nil {
		return err
	}
	if _, fileErr := os.Stat(LdapServersFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(LdapServersFileName, client.GetLdapServers); err != nil {
			return err
		}
	}
	if _, fileErr := os.Stat(SamlIdpFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(SamlIdpFileName, client.GetSamlIdentityProviders); err != nil {
			return err
		}
	}
	return nil
}

func fetchTeamsData(client *SASTClient, exporter *Export) error {
	log.Info().Msg("collecting teams")
	if err := exporter.AddFileWithDataSource(TeamsFileName, client.GetTeams); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(LdapTeamMappingsFileName, client.GetLdapTeamMappings); err != nil {
		return err
	}
	if err := exporter.AddFileWithDataSource(SamlTeamMappingsFileName, client.GetSamlTeamMappings); err != nil {
		return err
	}
	if _, fileErr := os.Stat(LdapServersFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(LdapServersFileName, client.GetLdapServers); err != nil {
			return err
		}
	}
	if _, fileErr := os.Stat(SamlIdpFileName); os.IsNotExist(fileErr) {
		if err := exporter.AddFileWithDataSource(SamlIdpFileName, client.GetSamlIdentityProviders); err != nil {
			return err
		}
	}
	return nil
}

func fetchResultsData(client *SASTClient, exporter *Export, resultsProjectActiveSince int) (err error) {
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
		go consumeReports(client, exporter, consumerID, reportJobs, reportConsumeOutputs)
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

func getTriagedScans(client *SASTClient, fromDate string) ([]TriagedScan, error) {
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
				continue
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
			ReportType: ScanReportTypeXML,
		}
	}
	close(reportJobs)
}

func consumeReports(client *SASTClient, exporter *Export, worker int, reportJobs <-chan ReportJob, done chan<- ReportConsumeOutput) {
	for reportJob := range reportJobs {
		report, reportErr := client.GetScanReport(reportJob.ScanID, reportJob.ReportType)
		if reportErr != nil {
			log.Debug().Err(reportErr).
				Int("ProjectID", reportJob.ProjectID).
				Int("ScanID", reportJob.ScanID).
				Int("worker", worker).
				Msg("failed fetching report")
			done <- ReportConsumeOutput{Err: reportErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
			continue
		}
		exportErr := exporter.AddFile(fmt.Sprintf(scansFileName, reportJob.ProjectID), report)
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
