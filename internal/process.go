package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/sliceutils"

	"github.com/rs/zerolog/log"
)

const (
	createdStatus = "Created"

	usersExportOption   = "users"
	teamsExportOption   = "teams"
	resultsExportOption = "results"

	reportType    = "XML"
	scansFileName = "%d.xml"
)

func RunExport(args *Args) {
	isDebug = args.Debug
	var selectedExportOptions []string
	consumerCount := GetNumCPU()

	log.Debug().
		Str("url", args.URL).
		Str("export", args.Export).
		Int("resultsProjectActiveSince", args.ResultsProjectActiveSince).
		Bool("debug", args.Debug).
		Int("consumers", consumerCount).
		Msg("starting export")

	// create api client and authenticate
	client, err := NewSASTClient(args.URL, &http.Client{})
	if err != nil {
		log.Error().Err(err)
		panic(err)
	}

	log.Info().Msg("connecting to SAST")
	if authErr := client.Authenticate(args.Username, args.Password); authErr != nil {
		log.Error().Err(authErr)
		panic(authErr)
	}

	if args.Export != "" {
		selectedExportOptions = strings.Split(args.Export, ",")
	} else {
		selectedExportOptions = []string{usersExportOption, resultsExportOption, teamsExportOption}
	}
	args.Export = strings.Join(selectedExportOptions, ",")

	log.Debug().
		Str("exportOptions", args.Export).
		Msgf("parsed export options")

	// start export
	log.Info().Msg("collecting data from SAST")

	exportValues, exportCreateErr := CreateExport(args.ProductName)
	if exportCreateErr != nil {
		log.Error().Err(exportCreateErr)
		panic(exportCreateErr)
	}

	if !isDebug {
		defer func(exportValues *Export) {
			cleanErr := exportValues.Clean()
			if cleanErr != nil {
				log.Error().Err(cleanErr).Msg("error cleaning export temporary folder")
			}
		}(&exportValues)
	}

	availableExportOptions := []string{usersExportOption, teamsExportOption, resultsExportOption}

	for _, exportOption := range availableExportOptions {
		if sliceutils.Contains(exportOption, sliceutils.ConvertStringToInterface(selectedExportOptions)) {
			log.Info().Msgf("collecting %s", exportOption)
			switch exportOption {
			case usersExportOption:
				if err = fetchUsersData(client); err != nil {
					log.Error().Err(err)
					panic(err)
				}
			case teamsExportOption:
				if err = fetchTeamsData(client); err != nil {
					log.Error().Err(err)
					panic(err)
				}
			case resultsExportOption:
				if err = fetchResultsData(client, args); err != nil {
					log.Error().Err(err)
					panic(err)
				}
			}
		}
	}

	log.Info().Msg("exporting collected data")

	exportFileName, exportErr := ExportResultsToFile(args, &exportValues)
	if exportErr != nil {
		log.Error().Err(exportErr).Msg("error exporting collected data")
	}

	log.Info().Msgf("export completed to %s", *exportFileName)
}

func ExportResultsToFile(args *Args, exportValues *Export) (*string, error) {

	if strings.Contains(args.Export, usersExportOption) {
		if exportErr := exportValues.AddFile(UsersFileName, usersData); exportErr != nil {
			return nil, exportErr
		}

		if exportErr := exportValues.AddFile(RolesFileName, rolesData); exportErr != nil {
			return nil, exportErr
		}

		if exportErr := exportValues.AddFile(LdapRoleMappingsFileName, ldapRolesData); exportErr != nil {
			return nil, exportErr
		}

		if exportErr := exportValues.AddFile(SamlRoleMappingsFileName, samlRolesData); exportErr != nil {
			return nil, exportErr
		}
	}

	if strings.Contains(args.Export, resultsExportOption) {
		for _, res := range exportData {
			if exportErr := exportValues.AddFile(res.FileName, res.Data); exportErr != nil {
				return nil, exportErr
			}
		}
	}

	if strings.Contains(args.Export, teamsExportOption) {
		if exportErr := exportValues.AddFile(TeamsFileName, teamsData); exportErr != nil {
			return nil, exportErr
		}

		if exportErr := exportValues.AddFile(LdapTeamMappingsFileName, ldapTeamsData); exportErr != nil {
			return nil, exportErr
		}

		if exportErr := exportValues.AddFile(SamlTeamMappingsFileName, samlTeamsData); exportErr != nil {
			return nil, exportErr
		}
	}

	if strings.Contains(args.Export, usersExportOption) || strings.Contains(args.Export, teamsExportOption) {
		if exportErr := exportValues.AddFile(LdapServersFileName, ldapServersData); exportErr != nil {
			return nil, exportErr
		}
		if exportErr := exportValues.AddFile(SamlIdpFileName, samlIDpsData); exportErr != nil {
			return nil, exportErr
		}
	}

	// create export package
	if isDebug {
		exec.Command(`explorer`, `/select,`, exportValues.TmpDir).Run()
		return &exportValues.TmpDir, nil
	}

	exportFileName, exportErr := exportValues.CreateExportPackage(args.ProductName, args.OutputPath)
	if exportErr != nil {
		return nil, exportErr
	}
	return &exportFileName, exportErr
}

func fetchUsersData(c *SASTClient) error {
	var err error
	usersData, err = c.GetUsers()
	if err != nil {
		return err
	}
	rolesData, err = c.GetRoles()
	if err != nil {
		return err
	}
	ldapRolesData, err = c.GetLdapRoleMappings()
	if err != nil {
		return err
	}
	samlRolesData, err = c.GetSamlRoleMappings()
	if err != nil {
		return err
	}
	ldapServersData, err = c.GetLdapServers()
	if err != nil {
		return err
	}
	samlIDpsData, err = c.GetSamlIdentityProviders()
	if err != nil {
		return err
	}
	return nil
}

func fetchTeamsData(client *SASTClient) error {
	var err error
	teamsData, err = client.GetTeams()
	if err != nil {
		return err
	}
	ldapTeamsData, err = client.GetLdapTeamMappings()
	if err != nil {
		return err
	}
	samlTeamsData, err = client.GetSamlTeamMappings()
	if err != nil {
		return err
	}
	ldapServersData, err = client.GetLdapServers()
	if err != nil {
		return err
	}
	samlIDpsData, err = client.GetSamlIdentityProviders()
	if err != nil {
		return err
	}
	return nil
}

func fetchResultsData(client *SASTClient, args *Args) (err error) {
	consumerCount := GetNumCPU()
	reportJobs := make(chan ReportJob)

	fromDate := GetDateFromDays(args.ResultsProjectActiveSince)

	log.Debug().
		Int("consumers", consumerCount).
		Str("startDate", fromDate).
		Msg("collecting results")

	var scans LastTriagedResponse
	dataScansOut, errGetLast := client.GetLastTriagedScanData(fromDate)
	if errGetLast != nil {
		return errGetLast
	}

	if errScansSheriff := json.Unmarshal(dataScansOut, &scans); errScansSheriff != nil {
		return errScansSheriff
	}

	log.Debug().
		Int("count", len(scans.Value)).
		Msg("triaged scans collected")

	scansList := convertTriagedScansResponseToLastScansList(scans)

	log.Debug().
		Int("count", len(scansList)).
		Str("scans", fmt.Sprintf("%v", scansList)).
		Msg("last scans by project")

	go produceReports(reportJobs, scansList)

	resultsCount := len(scansList)
	consumeErrors := make(chan error, resultsCount)

	for consumerID := 1; consumerID <= consumerCount; consumerID++ {
		go consumeReports(client, consumerID, reportJobs, consumeErrors)
	}

	for i := 0; i < resultsCount; i++ {
		consumeErr := <-consumeErrors
		resultIndex := i + 1
		if consumeErr == nil {
			log.Info().Msgf("collected result %d/%d", resultIndex, resultsCount)
		} else {
			log.Error().Err(consumeErr).Msgf("failed collecting result %d/%d", resultIndex, resultsCount)
		}
	}

	return nil
}

func produceReports(reportJobs chan<- ReportJob, scans []LastTriagedScanProducer) {
	for _, scan := range scans {
		reportJobs <- ReportJob{
			ProjectID:  scan.ProjectID,
			ScanID:     scan.ScanID,
			ReportType: reportType,
		}
	}
	close(reportJobs)
}

func consumeReports(client *SASTClient, worker int, reportJobs <-chan ReportJob, done chan<- error) {
	sleep := 2 * time.Second // default first time waiting, will increase in every loop
	retryAttempts := 4
	for reportJob := range reportJobs {
		logger := log.With().
			Int("projectID", reportJob.ProjectID).
			Int("scanID", reportJob.ScanID).
			Int("worker", worker).
			Logger()

		logger.Debug().Msg("consuming report job")

		// create the report
		reportBody := &ReportRequest{
			ReportType: reportJob.ReportType,
			ScanID:     reportJob.ScanID,
		}

		reportJSON, marshalErr := json.Marshal(reportBody)
		if marshalErr != nil {
			logger.Debug().
				Err(marshalErr).
				Str("reportBody", fmt.Sprintf("%v", reportBody)).
				Msg("failed marshalling report body")
			done <- marshalErr
			continue
		}
		body := bytes.NewBuffer(reportJSON)

		dataReportOut, errCreate := client.PostReportID(body)
		if errCreate != nil {
			logger.Debug().
				Err(errCreate).
				Str("reportBody", fmt.Sprintf("%v", reportBody)).
				Msg("failed creating report")
			done <- errCreate
			continue
		}

		var report ReportResponse
		errReportsSheriff := json.Unmarshal(dataReportOut, &report)
		if errReportsSheriff != nil {
			logger.Debug().
				Err(errReportsSheriff).
				Str("response", fmt.Sprintf("%v", string(dataReportOut))).
				Msg("failed unmarshalling report response")
			done <- errReportsSheriff
			continue
		}

		// monitor status and fetch
		status, errDoStatusReq := client.GetReportStatusResponse(report)
		if errDoStatusReq != nil {
			logger.Debug().Err(errDoStatusReq).Msg("failed getting report status")
			done <- errDoStatusReq
			continue
		}
		if status.Status.Value == createdStatus {
			err := fetchReportData(client, report.ReportID, reportJob.ProjectID)
			if err == nil {
				logger.Debug().Msg("report created")
			} else {
				logger.Debug().Err(err).Msg("failed getting report")
				done <- err
				continue
			}
		} else {
			err := retryGetReport(client, retryAttempts, report.ReportID, reportJob.ProjectID, sleep, report)
			if err != nil {
				logger.Debug().Err(err).Msg("failed report fetch retry")
				done <- err
				continue
			}
		}
		done <- nil
	}
}

func fetchReportData(client *SASTClient, reportID, projectID int) error {
	finalResultOut, errGetResult := client.GetReportResult(reportID)
	if errGetResult != nil {
		return errGetResult
	}

	exportData = append(exportData, ExportData{
		FileName: fmt.Sprintf(scansFileName, projectID),
		Data:     finalResultOut,
	})

	return nil
}

func retryGetReport(client *SASTClient, totalAttempts, reportID, projectID int, sleep time.Duration, response ReportResponse) error {
	state := true
	attempt := 2
	var status *StatusResponse
	var errDoStatusReq error
	for state {
		log.Debug().
			Int("attempt", attempt).
			Int("totalAttempts", totalAttempts).
			Int("projectID", projectID).
			Int("reportID", reportID).
			Str("sleep", sleep.String()).
			Msg("checking report create state")
		time.Sleep(sleep)
		sleep *= 2
		status, errDoStatusReq = client.GetReportStatusResponse(response)
		if errDoStatusReq != nil {
			return errDoStatusReq
		}

		state = status.Status.Value != createdStatus

		// Code to repeatedly execute until we have create status
		if status.Status.Value == createdStatus {
			state = false
			errReportData := fetchReportData(client, reportID, projectID)
			if errReportData != nil {
				return errReportData
			}
		}

		attempt++
		if attempt > totalAttempts {
			return fmt.Errorf("project %d report %d not ready after %d attempts", projectID, reportID, totalAttempts)
		}
	}
	return nil
}
