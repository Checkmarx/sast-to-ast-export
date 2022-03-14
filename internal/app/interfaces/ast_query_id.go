package interfaces

type ASTQueryIDProvider interface {
	GetQueryID(language, name, group string) (string, error)
}

type ASTQuery struct {
	Language, Name, Group, QueryID string
}
