package soap

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/rest"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	errRequestMarshalFailed    = "could not marshal request"
	errResponseUnmarshalFailed = "could not unmarshal response"
	errSoapCallFailed          = "SOAP call failed"
	errCannotGetQueryList      = "Cannot get Query list"
	errCannotGetPresetDetail   = "Cannot get preset detail %d"
)

type Adapter interface {
	GetSourcesByScanID(scanID string, filenames []string) (*GetSourcesByScanIDResponse, error)
	GetResultPathsForQuery(scanID string, queryID string) (*GetResultPathsForQueryResponse, error)
	GetQueryCollection() (*GetQueryCollectionResponse, error)
	GetPresetDetails(ID int) (*GetPresetDetailsResponse, error)
}

type Client struct {
	requestURL string
	httpClient rest.RetryableHTTPAdapter
	authToken  *rest.AccessToken
}

func NewClient(baseURL string, authToken *rest.AccessToken, adapter rest.RetryableHTTPAdapter) *Client {
	return &Client{
		requestURL: fmt.Sprintf("%s/Cxwebinterface/Portal/CxWebService.asmx", baseURL),
		authToken:  authToken,
		httpClient: adapter,
	}
}

func (e *Client) GetSourcesByScanID(scanID string, filenames []string) (*GetSourcesByScanIDResponse, error) {
	var fixedFilenames []string
	for _, filename := range filenames {
		if string(filename[0]) != "/" {
			filename = "/" + filename
		}
		fixedFilenames = append(fixedFilenames, filename)
	}
	requestBytes, marshalErr := xml.Marshal(GetSourcesByScanIDRequest{
		ScanID:          scanID,
		FilesToRetrieve: GetSourcesFilesToRetrieve{Strings: fixedFilenames},
	})
	if marshalErr != nil {
		return nil, errors.Wrap(marshalErr, errRequestMarshalFailed)
	}
	envelope, callErr := e.call("GetSourcesByScanID", requestBytes)
	if callErr != nil {
		return nil, errors.Wrap(callErr, errSoapCallFailed)
	}
	var response GetSourcesByScanIDResponse
	unmarshalErr := xml.Unmarshal(envelope.Body.Contents, &response)
	if unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, errResponseUnmarshalFailed)
	}
	return &response, nil
}

func (e *Client) GetResultPathsForQuery(scanID, queryID string) (*GetResultPathsForQueryResponse, error) {
	requestBytes, requestMarshalErr := xml.Marshal(GetResultPathsForQueryRequest{
		QueryID: queryID,
		ScanID:  scanID,
	})
	if requestMarshalErr != nil {
		return nil, errors.Wrap(requestMarshalErr, errRequestMarshalFailed)
	}
	envelope, callErr := e.call("GetResultPathsForQuery", requestBytes)
	if callErr != nil {
		return nil, errors.Wrap(callErr, errSoapCallFailed)
	}
	var response GetResultPathsForQueryResponse
	unmarshalErr := xml.Unmarshal(envelope.Body.Contents, &response)
	if unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, errResponseUnmarshalFailed)
	}
	if !response.GetResultPathsForQueryResult.IsSuccessful {
		return nil, fmt.Errorf("%s: %s", errSoapCallFailed, response.GetResultPathsForQueryResult.ErrorMessage)
	}
	return &response, nil
}

func (e *Client) GetQueryCollection() (*GetQueryCollectionResponse, error) {
	requestBytes, requestMarshalErr := xml.Marshal(GetQueryCollectionRequest{})
	if requestMarshalErr != nil {
		return nil, errors.Wrap(requestMarshalErr, errRequestMarshalFailed)
	}
	envelope, callErr := e.call("GetQueryCollection", requestBytes)
	if callErr != nil {
		return nil, errors.Wrap(callErr, errSoapCallFailed)
	}
	var response GetQueryCollectionResponse
	unmarshalErr := xml.Unmarshal(envelope.Body.Contents, &response)
	if unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, errResponseUnmarshalFailed)
	}
	if !response.GetQueryCollectionResult.IsSuccessful {
		return nil, fmt.Errorf("%s: %s", errSoapCallFailed, errCannotGetQueryList)
	}
	return &response, nil
}

func (e *Client) GetPresetDetails(id int) (*GetPresetDetailsResponse, error) {
	requestBytes, requestMarshalErr := xml.Marshal(GetPresetDetailsRequest{Id: id})
	if requestMarshalErr != nil {
		return nil, errors.Wrap(requestMarshalErr, errRequestMarshalFailed)
	}
	envelope, callErr := e.call("GetPresetDetails", requestBytes)
	if callErr != nil {
		return nil, errors.Wrap(callErr, errSoapCallFailed)
	}
	var response GetPresetDetailsResponse
	unmarshalErr := xml.Unmarshal(envelope.Body.Contents, &response)
	if unmarshalErr != nil {
		return nil, errors.Wrap(unmarshalErr, errResponseUnmarshalFailed)
	}
	if !response.GetPresetDetailsResult.IsSuccessful {
		return nil, fmt.Errorf("%s: "+errCannotGetPresetDetail, errSoapCallFailed, id)
	}
	return &response, nil
}

func (e *Client) call(soapAction string, requestBytes []byte) (*Envelope, error) {
	body := fmt.Sprintf(`
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:chec="http://Checkmarx.com">
   <soap:Header/>
   <soap:Body>
      %s
   </soap:Body>
</soap:Envelope>
`, string(requestBytes))
	req, reqErr := retryablehttp.NewRequest("POST", e.requestURL, strings.NewReader(body))
	if reqErr != nil {
		return nil, errors.Wrap(reqErr, "could not create request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", e.authToken.TokenType, e.authToken.AccessToken))
	req.Header.Add("Content-type", fmt.Sprintf("application/soap+xml;charset=UTF-8;action=http://Checkmarx.com/%s", soapAction))

	resp, doErr := e.httpClient.Do(req)
	if doErr != nil {
		return nil, errors.Wrap(doErr, "request failed")
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Msg("failed closing SOAP response body")
		}
	}()
	var envelope Envelope
	responseBody, responseBodyErr := io.ReadAll(resp.Body)
	if responseBodyErr != nil {
		return nil, errors.Wrap(responseBodyErr, "could not read response")
	}
	envelopeUnmarshalErr := xml.Unmarshal(responseBody, &envelope)
	if envelopeUnmarshalErr != nil {
		return nil, errors.Wrap(envelopeUnmarshalErr, "could not unmarshal envelope")
	}
	return &envelope, nil
}
