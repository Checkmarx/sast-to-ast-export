package astquery

import (
	"fmt"
	"hash/fnv"
	"strconv"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
)

const (
	queryIDIntBase       = 10
	notCustomPackageType = "Cx"
)

type Provider struct {
	queryProvider interfaces.QueriesRepo
}

func NewProvider(queryProvider interfaces.QueriesRepo) (*Provider, error) {
	return &Provider{
		queryProvider: queryProvider,
	}, nil
}

func (e *Provider) GetQueryID(language, name, group string) (string, error) {
	sourcePath := fmt.Sprintf("queries/%s/%s/%s/%s.cs", language, group, name, name)
	queryID, queryIDErr := hash(sourcePath)
	if queryIDErr != nil {
		return "", queryIDErr
	}
	return strconv.FormatUint(queryID, queryIDIntBase), nil
}

func (e *Provider) GetCustomQueriesList() (*soap.GetQueryCollectionResponse, error) {
	var output soap.GetQueryCollectionResponse
	queryResponse, err := e.queryProvider.GetQueriesList()
	if err != nil {
		return nil, err
	}

	output.GetQueryCollectionResult.IsSuccessful = true
	output.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup = []soap.CxWSQueryGroup{}

	for _, v := range queryResponse.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup {
		if v.PackageType != notCustomPackageType {
			output.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup =
				append(output.GetQueryCollectionResult.QueryGroups.CxWSQueryGroup, v)
		}
	}

	return &output, nil
}

func hash(s string) (uint64, error) {
	h := fnv.New64()
	_, err := h.Write([]byte(s))
	return h.Sum64(), err
}
