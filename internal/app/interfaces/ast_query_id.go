package interfaces

type ASTQueryIDProvider interface {
	GetQueryID(language, name, group, sastQueryID string) (string, error)
}

type ASTQuery struct {
	Language, Name, Group, QueryID string
}
