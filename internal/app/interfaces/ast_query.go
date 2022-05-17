package interfaces

import "github.com/checkmarxDev/ast-sast-export/internal/integration/soap"

type ASTQueryProvider interface {
	GetQueryID(language, name, group string) (string, error)
	GetCustomQueriesList() (*soap.GetQueryCollectionResponse, error)
}
