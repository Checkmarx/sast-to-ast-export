package method_line

import (
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	mock_integration_soap "github.com/checkmarxDev/ast-sast-export/test/mocks/integration/soap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	scanID       = "100000"
	queryID      = "3000"
	soapResponse = soap.GetResultPathsForQueryResponse{
		GetResultPathsForQueryResult: soap.GetResultPathsForQueryResult{
			Paths: soap.Paths{
				Paths: []soap.ResultPath{
					{
						PathID: "3",
						Node: soap.Node{
							Nodes: []soap.ResultPathNode{
								{MethodLine: "1"},
								{MethodLine: "2"},
								{MethodLine: "3"},
							},
						},
					},
					{
						PathID: "4",
						Node: soap.Node{
							Nodes: []soap.ResultPathNode{{MethodLine: "10"}, {MethodLine: "20"}, {MethodLine: "30"}},
						},
					},
				},
			},
		},
	}
)

func TestRepo_GetMethodLines(t *testing.T) {
	pathID := "3"
	ctrl := gomock.NewController(t)
	soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
	soapClientMock.EXPECT().GetResultPathsForQuery(scanID, queryID).Return(&soapResponse, nil)
	instance := NewRepo(soapClientMock)

	result, err := instance.GetMethodLines(scanID, queryID, pathID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3"}, result)
}

func TestRepo_GetMethodLinesByPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
	soapClientMock.EXPECT().GetResultPathsForQuery(scanID, queryID).Return(&soapResponse, nil)
	instance := NewRepo(soapClientMock)

	result, err := instance.GetMethodLinesByPath(scanID, queryID)
	assert.NoError(t, err)
	expected := map[string][]string{
		"3": {"1", "2", "3"},
		"4": {"10", "20", "30"},
	}
	assert.Equal(t, expected, result)
}
