package method_line

import (
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	mock_integration_soap "github.com/checkmarxDev/ast-sast-export/test/mocks/integration/soap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRepo_GetMethodLines(t *testing.T) {
	scanID := "100000"
	queryID := "3000"
	pathID := "3"
	soapResponse := soap.GetResultPathsForQueryResponse{
		GetResultPathsForQueryResult: soap.GetResultPathsForQueryResult{
			Paths: soap.Paths{
				Paths: []soap.ResultPath{
					{
						PathID: pathID,
						Node: soap.Node{
							Nodes: []soap.ResultPathNode{
								{MethodLine: "1"},
								{MethodLine: "2"},
								{MethodLine: "3"},
							},
						},
					},
					{
						PathID: "300",
						Node: soap.Node{
							Nodes: []soap.ResultPathNode{{MethodLine: "10"}, {MethodLine: "20"}, {MethodLine: "30"}},
						},
					},
				},
			},
		},
	}
	ctrl := gomock.NewController(t)
	soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
	soapClientMock.EXPECT().GetResultPathsForQuery(scanID, queryID).Return(&soapResponse, nil)
	instance := NewRepo(soapClientMock)

	result, err := instance.GetMethodLines(scanID, queryID, pathID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3"}, result)
}
