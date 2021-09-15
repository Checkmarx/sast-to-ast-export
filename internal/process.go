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

func RunExport(args Args) {
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

	log.Info().Msg("logging into SAST")
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

	log.Debug().Msgf("parsed export options: %s", args.Export)

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

	log.Info().Str("exportFile", *exportFileName).Msg("export completed")
}

func (c *SASTClient) produceReports(reports chan<- ReportConsumer, scans []LastTriagedScanProducer) error {
	for _, scan := range scans {
		reportBody := &ReportRequest{
			ReportType: reportType,
			ScanID:     scan.ScanID,
		}

		reportJSON, marshalErr := json.Marshal(reportBody)
		if marshalErr != nil {
			return marshalErr
		}
		body := bytes.NewBuffer(reportJSON)

		dataReportOut, errCreate := c.PostReportID(body)
		if errCreate != nil {
			return errCreate
		}

		var report ReportResponse
		errReportsSheriff := json.Unmarshal(dataReportOut, &report)
		if errReportsSheriff != nil {
			return errReportsSheriff
		}

		// add project and report id to producer list call
		reports <- ReportConsumer{
			ProjectId:      scan.ProjectID,
			ReportId:       report.ReportID,
			ReportResponse: report,
		}
	}
	close(reports)
	return nil
}

func (c *SASTClient) consumeReports(worker int, reports <-chan ReportConsumer, done chan<- bool) (err error) {
	sleep := 2 * time.Second // default first time waiting, will increase in every loop
	retryAttempts := 4
	for rep := range reports {
		log.Debug().Msgf("reportID %v is consumed by report worker %v.", rep.ReportId, worker)

		status, errDoStatusReq := GetReportStatusResponse(c, rep.ReportResponse)
		if errDoStatusReq != nil {
			return errDoStatusReq
		}

		if status.Status.Value == createdStatus {
			err = c.GetReportData(rep.ReportId, rep.ProjectId)
			if err != nil {
				return err
			}
		} else {
			err = c.retryGetReport(retryAttempts, rep.ReportId, rep.ProjectId, sleep, rep.ReportResponse, status)
			if err != nil {
				return err
			}
		}
	}
	done <- true
	return nil
}

func (c *SASTClient) GetScanDataResponse(args Args) (err error) {
	consumerCount := GetNumCPU()
	reports := make(chan ReportConsumer)
	done := make(chan bool, consumerCount)

	fromDate := GetDateFromDays(args.ResultsProjectActiveSince)

	log.Debug().
		Int("consumers", consumerCount).
		Str("startDate", fromDate).
		Msg("collecting results")

	var scans LastTriagedResponse
	dataScansOut, errGetLast := c.GetLastTriagedScanData(fromDate)
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

	go func() {
		err = c.produceReports(reports, scansList)
		if err != nil {
			log.Error().Err(err).Msg("error producing reports")
		}
	}()

	for i := 1; i <= consumerCount; i++ {
		consumerID := i
		go func() {
			err = c.consumeReports(consumerID, reports, done)
			if err != nil {
				log.Error().Err(err).Msg("error consuming reports")
			}
		}()
	}

	for i := 1; i <= consumerCount; i++ {
		<-done
	}

	return nil
}

func ExportResultsToFile(args Args, exportValues *Export) (*string, error) {

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

func (c *SASTClient) retryGetReport(attempts, reportID, projectID int, sleep time.Duration, response ReportResponse, status *StatusResponse) (err error) {
	state := true
	var errDoStatusReq error
	for state {
		log.Debug().
			Int("attempts", attempts).
			Int("projectID", projectID).
			Int("reportID", reportID).
			Str("sleep", sleep.String()).
			Msg("checking report create state")
		time.Sleep(sleep)
		sleep *= 2
		status, errDoStatusReq = GetReportStatusResponse(c, response)
		if errDoStatusReq != nil {
			return errDoStatusReq
		}

		state = status.Status.Value != createdStatus

		// Code to repeatedly execute until we have create status
		if status.Status.Value == createdStatus {
			state = false
			errReportData := c.GetReportData(reportID, projectID)
			if errReportData != nil {
				return errReportData
			}
		}

		attempts--
		if attempts == 0 {
			return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
		}
	}
	return nil
}

func (c *SASTClient) GetReportData(reportID, projectID int) error {
	finalResultOut, errGetResult := c.GetReportResult(reportID)
	if errGetResult != nil {
		return errGetResult
	}

	exportData = append(exportData, ExportData{
		FileName: fmt.Sprintf(scansFileName, projectID),
		Data:     finalResultOut,
	})

	log.Debug().
		Int("reportID", reportID).
		Msg("report created")

	return nil
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

func fetchTeamsData(c *SASTClient) error {
	var err error
	teamsData, err = c.GetTeams()
	if err != nil {
		return err
	}
	ldapTeamsData, err = c.GetLdapTeamMappings()
	if err != nil {
		return err
	}
	samlTeamsData, err = c.GetSamlTeamMappings()
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

func fetchResultsData(c *SASTClient, args Args) error {
	err := c.GetScanDataResponse(args)
	return err
}
