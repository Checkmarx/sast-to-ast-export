package interfaces

type ASTQueryIDRepo interface {
	GetQueryID(language, name, group string) (string, error)
	GetAllQueryIDsByGroup(language, name string) ([]ASTQuery, error)
}

type ASTQuery struct {
	Language, Name, Group, QueryID string
}
