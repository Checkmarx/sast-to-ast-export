package interfaces

import "github.com/checkmarxDev/ast-sast-export/internal/app/querymapping"

type QueryMappingRepo interface {
	GetMapping() []querymapping.QueryMap
	AddQueryMapping(language, name, group, sastQueryID string) error
}
