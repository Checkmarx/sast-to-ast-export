package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	time "time"
)

const (
	ReportType    = "XML"
	ScansFileName = "%d.xml"
)

type Result []interface{}

var (
	mu                   sync.Mutex
	exportList           []string
	exportData           []ExportData
	usersData, teamsData []byte
)

type HTTPAdapter interface {
	Do(req *http.Request) (*http.Response, error)
}

type SASTClient struct {
	BaseURL string
	Adapter HTTPAdapter
	Token   *AccessToken
}

func NewSASTClient(baseURL string, adapter HTTPAdapter) (*SASTClient, error) {
	client := SASTClient{
		BaseURL: baseURL,
		Adapter: adapter,
	}
	return &client, nil
}

func (c *SASTClient) Authenticate(username, password string) error {
	req, err := CreateAccessTokenRequest(c.BaseURL, username, password)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		errClose := Body.Close()
		if errClose != nil {

		}
	}(resp.Body)
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	c.Token = &AccessToken{}
	return json.Unmarshal(responseBody, c.Token)
}

func (c *SASTClient) GetUsersResponseBody() ([]byte, error) {
	req, err := CreateGetUsersRequest(c.BaseURL, c.Token)
	if err != nil {
		panic(err)
	}

	resp, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		errClose := Body.Close()
		if errClose != nil {

		}
	}(resp.Body)

	return ioutil.ReadAll(resp.Body)
}

func (c *SASTClient) GetTeamsResponseBody() ([]byte, error) {
	req, err := CreateGetTeamsRequest(c.BaseURL, c.Token)
	if err != nil {
		return []byte{}, err
	}

	resp, err := c.doRequest(req, http.StatusOK)
	if err != nil {
		return []byte{}, err
	}
	defer func(Body io.ReadCloser) {
		errClose := Body.Close()
		if errClose != nil {

		}
	}(resp.Body)

	return ioutil.ReadAll(resp.Body)
}

func produce(export chan<- string) {
	for _, msg := range exportList {
		export <- msg
	}
	close(export)
}

func consume(worker int, exportValues *Export, productName, outputPath string, client *SASTClient, export <-chan string, finished chan<- bool) {
	//defer wg.Done()
	//mu.Lock()
	//defer mu.Unlock()

	var err error

	for ch := range export {
		//time.Sleep(2 * time.Second)
		fmt.Printf("%v is consumed by worker %v.\n", ch, worker)
		switch ch {
		case "users": // fetch users and save to export dir
			usersData, err = client.GetUsersResponseBody()
			if err != nil {
				panic(err)
			}
		case "results": // fetch scans and save to export dir
			errResults := client.GetScanDataResponseBody()
			if errResults != nil {
				panic(errResults)
			}
		case "teams": // fetch teams and save to export dir
			teamsData, err = client.GetTeamsResponseBody()
			if err != nil {
				panic(err)
			}
		}
	}

	finished <- true
}

func GetAllData(url, username, password, export, outputPath, productName string) {
	consumerCount := 3 // default 3 consumerCount, for the max groups ("users", "results", "teams")
	exports := make(chan string)
	finished := make(chan bool, consumerCount)
	// create api client and authenticate
	client, err := NewSASTClient(url, &http.Client{})
	if err != nil {
		panic(err)
	}
	if err2 := client.Authenticate(username, password); err2 != nil {
		panic(err2)
	}

	if export != "" {
		exportList = strings.Split(export, ",")
		optionsCount := len(exportList)
		if optionsCount > 0 {
			// reset the default 3 consumerCount, if is only to export "users" for example
			consumerCount = optionsCount
		}
	} else {
		exportList = []string{"users", "results", "teams"}
	}

	go produce(exports)

	// start export
	exportValues, errCreateExport := CreateExport(productName)
	if errCreateExport != nil {
		panic(errCreateExport)
	}

	for i := 1; i <= consumerCount; i++ {
		go func() {
			consume(i, &exportValues, productName, outputPath, client, exports, finished)
			if err != nil {
			}
		}()
	}
	for i := 1; i <= consumerCount; i++ {
		<-finished
	}

	if exportErr := exportValues.AddFile(UsersFileName, usersData); exportErr != nil {
		panic(exportErr)
	}

	for _, res := range exportData {
		if exportErr := exportValues.AddFile(res.FileName, res.Data); exportErr != nil {
			panic(exportErr)
		}
	}

	if exportErr := exportValues.AddFile(TeamsFileName, teamsData); exportErr != nil {
		panic(exportErr)
	}

	// create export package
	exportFileName, exportErr := exportValues.CreateExportPackage(productName, outputPath)
	if exportErr != nil {
		panic(exportErr)
	}

	defer func(exportValues *Export) {
		errClean := exportValues.Clean()
		if errClean != nil {

		}
	}(&exportValues)

	fmt.Printf("SAST data exported to %s\n", exportFileName)
}

func loadProjects(c *SASTClient) ([]int, error) {
	var projectIds []int
	req, err := GetListProjectsRequest(c.BaseURL, c.Token)
	if err != nil {
		return []int{}, err
	}

	dataOut, errDoRequest := c.doRequest(req, http.StatusOK)

	if dataOut.StatusCode == http.StatusOK {
		if errDoRequest != nil {
			return []int{}, errDoRequest
		}
		defer func(Body io.ReadCloser) {
			errClose := Body.Close()
			if errClose != nil {

			}
		}(dataOut.Body)
	}
	projectsUnm, errReadAll := ioutil.ReadAll(dataOut.Body)
	if errReadAll != nil {
		return []int{}, errReadAll
	}

	var projectsInter Result
	if errSheriff := json.Unmarshal(projectsUnm, &projectsInter); errSheriff != nil {
		return []int{}, errSheriff
	}

	for _, proj := range projectsInter {
		m := proj.(map[string]interface{})
		projectIds = append(projectIds, int(m["id"].(float64)))
		//fmt.Printf("ProjectId: %d\n", int(m["id"].(float64)))
	}

	return projectIds, nil
}

func produceReports(reports chan<- ReportConsumer, c *SASTClient, projectIds []int) error {
	fmt.Print("Start producer\n")

	fmt.Printf("Start producer with reports: %v\n", projectIds)

	for _, projectId := range projectIds {
		fmt.Print("Inside new loop\n")
		fmt.Printf("ProjectId: %d\n", projectId)

		lastScanReq, errGetLast := GetLastScanDataRequest(c.BaseURL, projectId, c.Token)
		if errGetLast != nil {
			return errGetLast
		}

		dataScansOut, errScansDoRequest := c.doRequest(lastScanReq, http.StatusOK)
		if dataScansOut != nil && dataScansOut.StatusCode == http.StatusOK {
			if errScansDoRequest != nil {
				return errScansDoRequest
			}
			defer func(Body io.ReadCloser) {
				errClose := Body.Close()
				if errClose != nil {

				}
			}(dataScansOut.Body)
		}

		var scans Result
		scansUnm, errReadScansAll := ioutil.ReadAll(dataScansOut.Body)
		if errReadScansAll != nil {
			return errReadScansAll
		}

		if errScansSheriff := json.Unmarshal(scansUnm, &scans); errScansSheriff != nil {
			return errScansSheriff
		}

		for _, scan := range scans {
			s := scan.(map[string]interface{})
			reportBody := &ReportRequest{
				ReportType: ReportType,
				ScanID:     int(s["id"].(float64)),
			}

			body := dataToJSONReader(reportBody)
			reportReq, errCreate := CreateRequest(http.MethodPost, fmt.Sprintf("%s/CxRestAPI/help/reports/sastScan", c.BaseURL), body, c.Token)
			if errCreate != nil {
				return errCreate
			}

			dataReportOut, errReportDoRequest := c.doRequest(reportReq, http.StatusAccepted)
			if dataReportOut != nil && dataReportOut.StatusCode == http.StatusAccepted {
				if errReportDoRequest != nil {
					return errReportDoRequest
				}
				defer func(Body io.ReadCloser) {
					errClose := Body.Close()
					if errClose != nil {

					}
				}(dataReportOut.Body)
			}

			var report ReportResponse
			reportUnm, errReadReportsAll := ioutil.ReadAll(dataReportOut.Body)
			if errReadReportsAll != nil {
				return errReadReportsAll
			}

			errReportsSheriff := json.Unmarshal(reportUnm, &report)
			if errReportsSheriff != nil {
				return errReportsSheriff
			}

			// add id to producer list call

			reports <- ReportConsumer{
				ProjectId:      projectId,
				ReportId:       report.ReportID,
				ReportResponse: report,
			}
		}
	}
	fmt.Print("End producer\n")
	close(reports)
	return nil
}

func consumeReports(worker int, c *SASTClient, reports <-chan ReportConsumer, done chan<- bool) error {
	fmt.Print("Init consumer\n")
	for rep := range reports {
		fmt.Printf("ReportId %v is consumed by worker %v.\n", rep.ReportId, worker)
		reportStatusReq, errGetStatus := GetReportIDStatusRequest(c.BaseURL, strconv.Itoa(rep.ReportId), c.Token)
		if errGetStatus != nil {
			return errGetStatus
		}

		dataStatusOut, errReportDoRequest := c.doRequest(reportStatusReq, http.StatusOK)

		if dataStatusOut != nil && dataStatusOut.StatusCode == http.StatusOK {
			if errReportDoRequest != nil {
				return errReportDoRequest
			}
			defer func(Body io.ReadCloser) {
				errClose := Body.Close()
				if errClose != nil {

				}
			}(dataStatusOut.Body)
		}

		status, errDoStatusReq := doStatusRequest(c, rep.ReportResponse)
		if errDoStatusReq != nil {
			return errDoStatusReq
		}

		state := true
		retryAttempts := 10
		fmt.Printf("state value before loop: %v\n", state)

		for state {
			fmt.Printf("state value in loop: %v\n", state)

			time.Sleep(2 * time.Second)
			status, errDoStatusReq = doStatusRequest(c, rep.ReportResponse)
			if errDoStatusReq != nil {
				return errDoStatusReq
			}

			state = status.Status.Value != "Created"

			// Code to repeatedly execute until we have create status
			if status.Status.Value == "Created" {
				state = false
				fmt.Printf("state value in loop: %v\n", state)

				fmt.Printf("ReportId: %v\n", strconv.Itoa(rep.ReportId))

				finalResultRequest, errGetResult := CreateRequest(http.MethodGet, fmt.Sprintf("%s/CxRestAPI/help/reports/sastScan/%s", c.BaseURL, strconv.Itoa(rep.ReportId)), nil, c.Token)
				if errGetResult != nil {
					return errGetResult
				}

				respFinalResult, errDoFinalRequest := c.doRequest(finalResultRequest, http.StatusOK)
				if errDoFinalRequest != nil {
					return errDoFinalRequest
				}
				defer func(Body io.ReadCloser) {
					errClose := Body.Close()
					if errClose != nil {

					}
				}(respFinalResult.Body)

				reportFinal, errReadFinalReports := ioutil.ReadAll(respFinalResult.Body)
				if errReadFinalReports != nil {
					return errReadFinalReports
				}

				/*if exportErr := exportValue.AddFile(fmt.Sprintf(ScansFileName, rep.ProjectId), reportFinal); exportErr != nil {
					return []ExportData{}, exportErr
				}*/

				exportData = append(exportData, ExportData{
					FileName: fmt.Sprintf(ScansFileName, rep.ProjectId),
					Data:     reportFinal,
				})

				//fmt.Printf("End creating final report: %v\n",  string(reportFinal))

				fmt.Printf("End creating final report with ReportId: %v\n", strconv.Itoa(rep.ReportId))
			}

			retryAttempts -= 1
			if retryAttempts == 0 {
				break
			}
		}
	}
	done <- true
	fmt.Print("End consumer\n")
	return nil
}

func (c *SASTClient) GetScanDataResponseBody() error {
	consumerCount := runtime.GOMAXPROCS(runtime.NumCPU())
	reports := make(chan ReportConsumer)
	done := make(chan bool, consumerCount)
	var err error

	projects, errLoadProjects := loadProjects(c)
	if errLoadProjects != nil {
		panic(err)
	}

	fmt.Print("After load projects\n")

	// producer call

	go func() {
		err = produceReports(reports, c, projects)
		if err != nil {
		}
	}()

	/// consumer call

	for i := 1; i <= consumerCount; i++ {
		go func() {
			err = consumeReports(i, c, reports, done)
			if err != nil {
			}
		}()
	}

	for i := 1; i <= consumerCount; i++ {
		<-done
	}

	return nil
}

func doStatusRequest(c *SASTClient, report ReportResponse) (StatusResponse, error) {
	reportStatusReq, errGetStatus := GetReportIDStatusRequest(c.BaseURL, strconv.Itoa(report.ReportID), c.Token)
	if errGetStatus != nil {
		return StatusResponse{}, errGetStatus
	}

	dataStatusOut, errReportDoRequest := c.doRequest(reportStatusReq, http.StatusOK)

	if dataStatusOut != nil && dataStatusOut.StatusCode == http.StatusOK {
		if errReportDoRequest != nil {
			return StatusResponse{}, errReportDoRequest
		}
		defer func(Body io.ReadCloser) {
			errClose := Body.Close()
			if errClose != nil {

			}
		}(dataStatusOut.Body)
	}

	var status StatusResponse
	statusUnm, errReadReportsAll := ioutil.ReadAll(dataStatusOut.Body)
	if errReadReportsAll != nil {
		return StatusResponse{}, errReadReportsAll
	}

	errStatusSheriff := json.Unmarshal(statusUnm, &status)
	if errStatusSheriff != nil {
		return StatusResponse{}, errStatusSheriff
	}

	return status, nil
}

func (c *SASTClient) doRequest(request *http.Request, expectStatusCode int) (*http.Response, error) {
	fmt.Printf("doRequest url: %s - method: %s\n", request.URL, request.Method)

	resp, err := c.Adapter.Do(request)

	if err != nil {
		return nil, err
	}

	fmt.Printf("doRequest status code: %d\n", resp.StatusCode)

	if resp.StatusCode != expectStatusCode {
		return nil, fmt.Errorf("invalid response: %v", resp)
	}

	return resp, nil
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			fmt.Printf("retrying after error: %s", err)
			time.Sleep(sleep)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func deferBody(response *http.Response) {
	if response != nil && response.StatusCode == http.StatusOK {
		defer func(Body io.ReadCloser) {
			errClose := Body.Close()
			if errClose != nil {

			}
		}(response.Body)
	}
}
