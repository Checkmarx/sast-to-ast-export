package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/checkmarxDev/ast-sast-export/internal/export"
	"github.com/checkmarxDev/ast-sast-export/internal/permissions"
	"github.com/checkmarxDev/ast-sast-export/internal/sliceutils"
	"github.com/dgrijalva/jwt-go"
	"github.com/rs/zerolog/log"
)

const (
	createdStatus = "Created"

	reportType    = "XML"
	scansFileName = "%d.xml"

	triagedScansPageLimit = 1000
)

var (
	exportData []ExportData
	usersData,
	rolesData,
	ldapRolesData,
	samlRolesData,
	teamsData,
	samlTeamsData,
	ldapTeamsData,
	samlIDpsData,
	ldapServersData []byte
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
		Int("resultsProjectActiveSince", args.ResultsProjectActiveSince).
		Bool("debug", args.Debug).
		Int("consumers", consumerCount).
		Msg("starting export")

	// create api client
	client, err := NewSASTClient(args.URL, &http.Client{})
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

	fetchErr := fetchSelectedData(client, args)
	if fetchErr != nil {
		log.Error().Err(err)
		panic(fmt.Errorf("fetch error - %s", err.Error()))
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
	exportOptions := sliceutils.ConvertStringToInterface(args.Export)
	if sliceutils.Contains(export.UsersOption, exportOptions) {
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

	if sliceutils.Contains(export.ResultsOption, exportOptions) {
		for _, res := range exportData {
			if exportErr := exportValues.AddFile(res.FileName, res.Data); exportErr != nil {
				return nil, exportErr
			}
		}
	}

	if sliceutils.Contains(export.TeamsOption, exportOptions) {
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

	if sliceutils.Contains(export.UsersOption, exportOptions) || sliceutils.Contains(export.TeamsOption, exportOptions) {
		if exportErr := exportValues.AddFile(LdapServersFileName, ldapServersData); exportErr != nil {
			return nil, exportErr
		}
		if exportErr := exportValues.AddFile(SamlIdpFileName, samlIDpsData); exportErr != nil {
			return nil, exportErr
		}
	}

	// create export package
	if args.Debug {
		exec.Command(`explorer`, `/select,`, exportValues.TmpDir).Run()
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
			log.Warn().Msgf("missing permission %s", description)
		}
		return fmt.Errorf("please add missing permissions to your SAST user")
	}
	return nil
}

func fetchSelectedData(client *SASTClient, args *Args) error {
	options := sliceutils.ConvertStringToInterface(args.Export)
	for _, exportOption := range export.GetOptions() {
		if sliceutils.Contains(exportOption, options) {
			log.Info().Msgf("collecting %s", exportOption)
			switch exportOption {
			case export.UsersOption:
				if err := fetchUsersData(client); err != nil {
					return err
				}
			case export.TeamsOption:
				if err := fetchTeamsData(client); err != nil {
					return err
				}
			case export.ResultsOption:
				if err := fetchResultsData(client, args.ResultsProjectActiveSince); err != nil {
					return err
				}
			}
		}
	}
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

func fetchResultsData(client *SASTClient, resultsProjectActiveSince int) (err error) {
	consumerCount := GetNumCPU()
	reportJobs := make(chan ReportJob)

	fromDate := GetDateFromDays(resultsProjectActiveSince, time.Now())

	// collect last triaged scan by project
	var lastTriagedScansByProject []TriagedScan
	triagedScansOffset := 0
	triagedScansLimit := triagedScansPageLimit

	for {
		var triagedScansPage TriagedScansResponse

		log.Debug().
			Str("fromDate", fromDate).
			Int("offset", triagedScansOffset).
			Int("limit", triagedScansLimit).
			Msg("fetching triaged scans")

		triagedScansResponse, errFetch := client.GetTriagedScansFromDate(fromDate, triagedScansOffset, triagedScansLimit)
		if errFetch != nil {
			return errFetch
		}

		if errUnmarshal := json.Unmarshal(triagedScansResponse, &triagedScansPage); errUnmarshal != nil {
			return errUnmarshal
		}

		if len(triagedScansPage.Value) == 0 {
			log.Debug().
				Str("fromDate", fromDate).
				Int("offset", triagedScansOffset).
				Int("limit", triagedScansLimit).
				Msg("finished fetching triaged scans")
			break
		}

		log.Debug().
			Str("fromDate", fromDate).
			Int("offset", triagedScansOffset).
			Int("limit", triagedScansLimit).
			Int("count", len(triagedScansPage.Value)).
			Int("responseSize", len(triagedScansResponse)).
			Msg("triaged scans fetched")

		scans := convertTriagedScansResponseToScansList(triagedScansPage)
		lastTriagedScansByProject = append(lastTriagedScansByProject, scans...)
		lastTriagedScansByProject = getLastScansByProject(lastTriagedScansByProject)

		triagedScansOffset += triagedScansLimit
	}

	log.Debug().
		Int("count", len(lastTriagedScansByProject)).
		Str("scans", fmt.Sprintf("%v", lastTriagedScansByProject)).
		Msg("last scans by project")

	// create and fetch report for each scan
	go produceReports(reportJobs, lastTriagedScansByProject)

	reportCount := len(lastTriagedScansByProject)
	reportConsumeOutputs := make(chan ReportConsumeOutput, reportCount)

	for consumerID := 1; consumerID <= consumerCount; consumerID++ {
		go consumeReports(client, consumerID, reportJobs, reportConsumeOutputs)
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
			log.Debug().
				Err(consumeOutput.Err).
				Int("projectID", consumeOutput.ProjectID).
				Int("scanID", consumeOutput.ScanID).
				Msg("failed collecting result")
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

func produceReports(reportJobs chan<- ReportJob, scans []TriagedScan) {
	for _, scan := range scans {
		reportJobs <- ReportJob{
			ProjectID:  scan.ProjectID,
			ScanID:     scan.ScanID,
			ReportType: reportType,
		}
	}
	close(reportJobs)
}

func consumeReports(client *SASTClient, worker int, reportJobs <-chan ReportJob, done chan<- ReportConsumeOutput) {
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
			done <- ReportConsumeOutput{Err: marshalErr, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
			continue
		}
		body := bytes.NewBuffer(reportJSON)

		dataReportOut, errCreate := client.PostReportID(body)
		if errCreate != nil {
			logger.Debug().
				Err(errCreate).
				Str("reportBody", fmt.Sprintf("%v", reportBody)).
				Msg("failed creating report")
			done <- ReportConsumeOutput{Err: errCreate, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
			continue
		}

		var report ReportResponse
		errReportsSheriff := json.Unmarshal(dataReportOut, &report)
		if errReportsSheriff != nil {
			logger.Debug().
				Err(errReportsSheriff).
				Str("response", fmt.Sprintf("%v", string(dataReportOut))).
				Msg("failed unmarshalling report response")
			done <- ReportConsumeOutput{Err: errReportsSheriff, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
			continue
		}

		// monitor status and fetch
		status, errDoStatusReq := client.GetReportStatusResponse(report)
		if errDoStatusReq != nil {
			logger.Debug().Err(errDoStatusReq).Msg("failed getting report status")
			done <- ReportConsumeOutput{Err: errDoStatusReq, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
			continue
		}
		if status.Status.Value == createdStatus {
			err := fetchReportData(client, report.ReportID, reportJob.ProjectID)
			if err == nil {
				logger.Debug().Msg("report created")
			} else {
				logger.Debug().Err(err).Msg("failed getting report")
				done <- ReportConsumeOutput{Err: err, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
				continue
			}
		} else {
			err := retryGetReport(client, retryAttempts, report.ReportID, reportJob.ProjectID, sleep, report)
			if err != nil {
				logger.Debug().Err(err).Msg("failed report fetch retry")
				done <- ReportConsumeOutput{Err: err, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
				continue
			}
		}
		done <- ReportConsumeOutput{Err: nil, ProjectID: reportJob.ProjectID, ScanID: reportJob.ScanID}
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
