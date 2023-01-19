package installation

import (
	"testing"

	mock_integration_soap "github.com/checkmarxDev/ast-sast-export/test/mocks/integration/soap"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	soapResponseSuccess = soap.GetInstallationSettingsResponse{
		GetInstallationSettingsResult: soap.GetInstallationSettingsResult{
			IsSuccesfull: "true",
			InstallationSettingsList: soap.InstallationSettingsList{
				InstallationSetting: []*soap.InstallationSetting{
					{
						Name:    "Checkmarx Engine Service",
						Version: "9.3.4.1111",
						Hotfix:  "Hotfix",
					},
					{
						Name:    "Checkmarx Queries Pack",
						Version: "9.3.4.5111",
						Hotfix:  "Hotfix",
					},
				},
			},
		},
	}
)

func TestRepo_GetInstallationSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
	soapClientMock.EXPECT().GetInstallationSettings().Return(&soapResponseSuccess, nil)
	instance := NewRepo(soapClientMock)

	result, err := instance.GetInstallationSettings()
	assert.NoError(t, err)
	assert.Equal(t, &soapResponseSuccess, result)
}
