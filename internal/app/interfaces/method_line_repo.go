package interfaces

type MethodLineRepo interface {
	GetMethodLines(scanID, queryID, pathID string) ([]string, error)
	GetMethodLinesByPath(scanID, queryID string) ([]*ResultPath, error)
}

type ResultPath struct {
	PathID      string
	MethodLines []string
}
