package astquery

import (
	"encoding/xml"

	"github.com/checkmarxDev/ast-sast-export/internal/app/common"
	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/checkmarxDev/ast-sast-export/internal/app/querymapping"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
)

const (
	notCustomPackageType = "Cx"
)

type (
	Provider struct {
		queryProvider interfaces.QueriesRepo
		mapping       []querymapping.QueryMap
	}
)

func NewProvider(queryProvider interfaces.QueriesRepo, queryMappingProvider interfaces.QueryMappingRepo) (*Provider, error) {
	return &Provider{
		queryProvider: queryProvider,
		mapping:       queryMappingProvider.GetMapping(),
	}, nil
}

func (e *Provider) GetQueryID(language, name, group, sastQueryID string) (string, error) {
	mappedAstID := e.getMappedID(sastQueryID)
	if mappedAstID != "" {
		return mappedAstID, nil
	}
	return common.GetAstQueryID(language, name, group)
}

func (e *Provider) GetCustomQueriesList() (*soap.GetQueryCollectionResponse, error) {
	var output soap.GetQueryCollectionResponse
	queryResponse, err := e.queryProvider.GetQueriesList()
	if err != nil {
		return nil, err
	}

	output.XMLName = xml.Name{Local: "GetQueryCollectionResponse"}
	output.GetQueryCollectionResult.IsSuccessful = true
	output.GetQueryCollectionResult.XMLName = xml.Name{Local: "GetQueryCollectionResult"}
	output.GetQueryCollectionResult.QueryGroups.XMLName = xml.Name{Local: "QueryGroups"}
	output.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup = []soap.CxWSQueryGroup{}

	for _, v := range queryResponse.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup {
		if v.PackageType != notCustomPackageType {
			output.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup =
				append(output.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup, v)
		}
	}

	return &output, nil
}

func (e *Provider) getMappedID(sastID string) string {
	for _, queryMap := range e.mapping {
		if queryMap.SastID == sastID {
			return queryMap.AstID
		}
	}
	return ""
}
