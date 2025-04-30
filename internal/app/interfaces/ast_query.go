package interfaces

import "github.com/checkmarxDev/ast-sast-export/internal/integration/soap"

type ASTQueryProvider interface {
	GetQueryID(language, name, group, sastQueryID string) (string, error)
	GetCustomQueriesList() (*soap.GetQueryCollectionResponse, error)
	GetCustomStatesList() (*soap.GetResultStateListResponse, error)
	GetStateMapping() (map[string]string, error)
	GetRawCustomStatesList() (*soap.GetResultStateListResponse, error)
}
