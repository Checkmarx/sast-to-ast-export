package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func GetNumCPU() int {
	numCpu := runtime.GOMAXPROCS(runtime.NumCPU())
	// Not allow more than 4 cpu's
	if numCpu > 4 {
		numCpu = 4
	}
	if isDebug {
		fmt.Printf("Core count: %v\n", numCpu)
	}
	return numCpu
}

func produce(export chan<- string) {
	for _, exp := range exportList {
		export <- exp
	}
	close(export)
}

func (c *SASTClient) consume(worker int, export <-chan string, finished chan<- bool) {
	var err error
	for ch := range export {
		if isDebug {
			fmt.Printf("%v is consumed by worker %v.\n", ch, worker)
		}
		switch ch {
		case Users: // fetch users and save to export dir
			usersData, err = c.GetUsers()
			if err != nil {
				panic(err)
			}
			rolesData, err = c.GetRoles()
			if err != nil {
				panic(err)
			}
			ldapRolesData, err = c.GetLdapRoleMappings()
			if err != nil {
				panic(err)
			}
			samlRolesData, err = c.GetSamlRoleMappings()
			if err != nil {
				panic(err)
			}
		case Results: // fetch scans and save to export dir
			errResults := c.GetScanDataResponse()
			if errResults != nil {
				panic(errResults)
			}
		case Teams: // fetch teams and save to export dir
			teamsData, err = c.GetTeams()
			if err != nil {
				panic(err)
			}
			ldapTeamsData, err = c.GetLdapTeamMappings()
			if err != nil {
				panic(err)
			}
			samlTeamsData, err = c.GetSamlTeamMappings()
			if err != nil {
				panic(err)
			}
		}

		switch ch {
		case Users, Teams: // fetch users | teams and save to export dir
			ldapServersData, err = c.GetLdapServers()
			if err != nil {
				panic(err)
			}

			samlIdpsData, err = c.GetSamlIdentityProviders()
			if err != nil {
				panic(err)
			}
		}
	}

	finished <- true
}

func RunExport(args Args) {
	isDebug = args.Debug
	consumerCount := 3 // default 3 consumerCount, for the max groups ("users", "results", "teams")
	exports := make(chan string)
	finished := make(chan bool, consumerCount)
	// create api client and authenticate
	client, err := NewSASTClient(args.Url, &http.Client{})
	if err != nil {
		panic(err)
	}
	if err2 := client.Authenticate(args.Username, args.Password); err2 != nil {
		panic(err2)
	}

	if args.Export != "" {
		exportList = strings.Split(args.Export, ",")
		optionsCount := len(exportList)
		if optionsCount > 0 {
			// reset the default 3 consumerCount, if is only to export "users" for example
			consumerCount = optionsCount
		}
	} else {
		exportList = []string{Users, Results, Teams}
	}

	go produce(exports)

	// start export
	exportValues, errCreateExport := CreateExport(args.ProductName)
	if errCreateExport != nil {
		panic(errCreateExport)
	}

	if !isDebug {
		defer func(exportValues *Export) {
			errClean := exportValues.Clean()
			if errClean != nil {

			}
		}(&exportValues)
	}

	for i := 1; i <= consumerCount; i++ {
		i := i
		go func() {
			client.consume(i, exports, finished)
			if err != nil {
			}
		}()
	}
	for i := 1; i <= consumerCount; i++ {
		<-finished
	}

	ExportResultsToFile(args, &exportValues)
}

func (c *SASTClient) produceReports(reports chan<- ReportConsumer, projectIds []int) error {
	if isDebug {
		fmt.Printf("producer with reports: %v\n", projectIds)
	}

	for _, projectId := range projectIds {
		dataScansOut, errGetLast := c.GetLastScanData(projectId, 1)
		if errGetLast != nil {
			return errGetLast
		}

		var scans Result

		if errScansSheriff := json.Unmarshal(dataScansOut, &scans); errScansSheriff != nil {
			return errScansSheriff
		}

		for _, scan := range scans {
			s := scan.(map[string]interface{})
			reportBody := &ReportRequest{
				ReportType: ReportType,
				ScanID:     int(s["id"].(float64)),
			}

			body := dataToJSONReader(reportBody)
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
				ProjectId:      projectId,
				ReportId:       report.ReportID,
				ReportResponse: report,
			}
		}
	}
	close(reports)
	return nil
}

func (c *SASTClient) consumeReports(worker int, reports <-chan ReportConsumer, done chan<- bool) error {
	sleep := 2 * time.Second // default first time waiting, will increase in every loop
	retryAttempts := 4
	for rep := range reports {
		if isDebug {
			fmt.Printf("ReportId %v is consumed by report worker %v.\n", rep.ReportId, worker)
		}

		status, errDoStatusReq := GetReportStatusResponse(c, rep.ReportResponse)
		if errDoStatusReq != nil {
			return errDoStatusReq
		}

		if status.Status.Value == "Created" {
			errDoStatusResult := c.GetReportData(rep.ReportId, rep.ProjectId)
			if errDoStatusResult != nil {
				return errDoStatusResult
			}
		} else {
			errDoRetry := c.retryGetReport(retryAttempts, rep.ReportId, rep.ProjectId, sleep, rep.ReportResponse, status)
			if errDoRetry != nil {
				return errDoRetry
			}
		}
	}
	done <- true
	return nil
}

func (c *SASTClient) GetScanDataResponse() error {
	consumerCount := GetNumCPU()
	reports := make(chan ReportConsumer)
	done := make(chan bool, consumerCount)
	var err error

	projects, errLoadProjects := c.GetAllProjects()
	if errLoadProjects != nil {
		panic(err)
	}

	go func() {
		err = c.produceReports(reports, projects)
		if err != nil {
		}
	}()

	for i := 1; i <= consumerCount; i++ {
		i := i
		go func() {
			err = c.consumeReports(i, reports, done)
			if err != nil {
			}
		}()
	}

	for i := 1; i <= consumerCount; i++ {
		<-done
	}

	return nil
}

func ExportResultsToFile(args Args, exportValues *Export) {

	if strings.Contains(args.Export, Users) {
		if exportErr := exportValues.AddFile(UsersFileName, usersData); exportErr != nil {
			panic(exportErr)
		}

		if exportErr := exportValues.AddFile(RolesFileName, rolesData); exportErr != nil {
			panic(exportErr)
		}

		if exportErr := exportValues.AddFile(LdapRoleMappingsFileName, ldapRolesData); exportErr != nil {
			panic(exportErr)
		}

		if exportErr := exportValues.AddFile(SamlRoleMappingsFileName, samlRolesData); exportErr != nil {
			panic(exportErr)
		}

	}

	if strings.Contains(args.Export, Results) {
		for _, res := range exportData {
			if exportErr := exportValues.AddFile(res.FileName, res.Data); exportErr != nil {
				panic(exportErr)
			}
		}
	}

	if strings.Contains(args.Export, Teams) {
		if exportErr := exportValues.AddFile(TeamsFileName, teamsData); exportErr != nil {
			panic(exportErr)
		}

		if exportErr := exportValues.AddFile(LdapTeamMappingsFileName, ldapTeamsData); exportErr != nil {
			panic(exportErr)
		}

		if exportErr := exportValues.AddFile(SamlTeamMappingsFileName, samlTeamsData); exportErr != nil {
			panic(exportErr)
		}
	}

	if strings.Contains(args.Export, Users) || strings.Contains(args.Export, Teams) {
		if exportErr := exportValues.AddFile(LdapServersFileName, ldapServersData); exportErr != nil {
			panic(exportErr)
		}
		if exportErr := exportValues.AddFile(SamlIdpFileName, samlIdpsData); exportErr != nil {
			panic(exportErr)
		}
	}

	// create export package
	if !isDebug {
		exportFileName, exportErr := exportValues.CreateExportPackage(args.ProductName, args.OutputPath)
		if exportErr != nil {
			panic(exportErr)
		}
		fmt.Printf("SAST data exported to %s\n", exportFileName)
	} else {
		fmt.Printf("Debug mode: SAST data exported to %s\n", exportValues.TmpDir)
		cmd := exec.Command(`explorer`, `/select,`, exportValues.TmpDir)
		errRun := cmd.Run()
		if errRun != nil {
		}
	}
}

func (c *SASTClient) retryGetReport(attempts, reportId, projectId int, sleep time.Duration, response ReportResponse, status StatusResponse) (err error) {
	state := true
	var errDoStatusReq error
	for state {
		time.Sleep(sleep)
		sleep *= 2
		status, errDoStatusReq = GetReportStatusResponse(c, response)
		if errDoStatusReq != nil {
			return errDoStatusReq
		}

		state = status.Status.Value != "Created"

		// Code to repeatedly execute until we have create status
		if status.Status.Value == "Created" {
			state = false
			errDoStatusResult := c.GetReportData(reportId, projectId)
			if errDoStatusResult != nil {
				return errDoStatusResult
			}
		}

		attempts -= 1
		if attempts == 0 {
			return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
		}
	}
	return nil
}

func (c *SASTClient) GetReportData(reportId, projectId int) error {
	finalResultOut, errGetResult := c.GetReportResult(reportId)
	if errGetResult != nil {
		return errGetResult
	}

	exportData = append(exportData, ExportData{
		FileName: fmt.Sprintf(ScansFileName, projectId),
		Data:     finalResultOut,
	})

	if isDebug {
		fmt.Printf("End creating final report with ReportId: %d\n", reportId)
	}
	return nil
}
