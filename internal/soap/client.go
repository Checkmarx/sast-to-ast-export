package soap

import "fmt"

type Adapter interface {
	Authenticate(username, password string) error
	GetSourcesByScanID(scanID string, filenames []string) (*GetSourcesByScanIDResponse, error)
	GetResultPathsForQuery(scanID string, queryID string) (*GetResultPathsForQueryResponse, error)
}

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (e *Client) Authenticate(username, password string) error {
	return fmt.Errorf("not implemented")
}

func (e *Client) GetSourcesByScanID(scanID string, filenames []string) (*GetSourcesByScanIDResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *Client) GetResultPathsForQuery(scanID, queryID string) (*GetResultPathsForQueryResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
