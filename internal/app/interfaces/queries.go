package interfaces

import "github.com/checkmarxDev/ast-sast-export/internal/integration/soap"

type QueriesRepo interface {
	GetQueriesList() (*soap.GetQueryCollectionResponse, error)
	GetCustomStatesList() (*soap.GetResultStateListResponse, error)
}
