package interfaces

type MethodLineRepo interface {
	GetMethodLines(scanID, queryID, pathID string) ([]string, error)
	GetMethodLinesByPath(scanID, queryID string) (map[string][]string, error)
}
