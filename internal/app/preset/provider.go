package preset

import (
	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"
	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
)

type Provider struct {
	presetProvider interfaces.PresetProvider
}

func NewProvider(presetProvider interfaces.PresetProvider) *Provider {
	return &Provider{
		presetProvider: presetProvider,
	}
}

func (e *Provider) GetPresetDetails(id int) (*soap.GetPresetDetailsResponse, error) {
	return e.presetProvider.GetPresetDetails(id)
}
