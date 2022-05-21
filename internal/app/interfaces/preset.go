package interfaces

import "github.com/checkmarxDev/ast-sast-export/internal/integration/soap"

type PresetProvider interface {
	GetPresetDetails(ID int) (*soap.GetPresetDetailsResponse, error)
}
