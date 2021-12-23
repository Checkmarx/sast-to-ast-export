package interfaces

type ASTQueryIDRepo interface {
	GetQueryID(language, name, group string) (string, error)
}
