package ast

import "fmt"

type QueryIDProvider interface {
	GetQueryID(language, name, group string) (string, error)
}

type QueryIDRepo struct{}

func NewQueryIDRepo() *QueryIDRepo {
	return &QueryIDRepo{}
}

func (e *QueryIDRepo) GetQueryID(language, name, group string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
