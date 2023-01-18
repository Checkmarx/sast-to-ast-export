package interfaces

import "github.com/checkmarxDev/ast-sast-export/internal/integration/soap"

type InstallationProvider interface {
	GetInstallationSettings() (*soap.GetInstallationSettingsResponse, error)
}
