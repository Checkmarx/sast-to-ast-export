package source_file

import (
	"fmt"
	"os"
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	mock_integration_soap "github.com/checkmarxDev/ast-sast-export/test/mocks/integration/soap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRepo_DownloadSourceFiles(t *testing.T) {
	scanID := "1000000"
	file1 := "project/folder1/file1.go"
	file2 := "project/folder1/file2.go"
	file3 := "project/folder2/file3.go"
	t.Run("success case", func(t *testing.T) {
		tmpDir := t.TempDir()
		filesToDownload := map[string]string{
			file1: fmt.Sprintf("%s/%s", tmpDir, file1),
			file2: fmt.Sprintf("%s/%s", tmpDir, file2),
			file3: fmt.Sprintf("%s/%s", tmpDir, file3),
		}
		fileSources := map[string]string{
			file1: "file1",
			file2: "file2",
			file3: "file3",
		}
		soapResponse := soap.GetSourcesByScanIDResponse{
			GetSourcesByScanIDResult: soap.GetSourcesByScanIDResult{
				CxWSResponseSourcesContent: soap.CxWSResponseSourcesContent{
					CxWSResponseSourceContents: []soap.CxWSResponseSourceContent{
						{Source: fileSources[file1]},
						{Source: fileSources[file2]},
						{Source: fileSources[file3]},
					},
				},
			},
		}
		getSourcesHandler := func(_ string, files []string) (*soap.GetSourcesByScanIDResponse, error) {
			assert.ElementsMatch(t, files, []string{file1, file2, file3})
			return &soapResponse, nil
		}

		ctrl := gomock.NewController(t)
		soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
		soapClientMock.EXPECT().GetSourcesByScanID(scanID, gomock.Any()).DoAndReturn(getSourcesHandler)
		instance := NewRepo(soapClientMock)

		err := instance.DownloadSourceFiles(scanID, filesToDownload)
		assert.NoError(t, err)

		assertFileExistWithContent(t, filesToDownload, fileSources)
	})
	t.Run("only downloads files that don't exist in local filesystem", func(t *testing.T) {
		tmpDir := t.TempDir()
		filesToDownload := map[string]string{
			file1: fmt.Sprintf("%s/%s", tmpDir, file1),
			file2: fmt.Sprintf("%s/%s", tmpDir, file2),
			file3: fmt.Sprintf("%s/%s", tmpDir, file3),
		}
		fileSources := map[string]string{
			file1: "file1",
			file2: "file2 already exists",
			file3: "file3",
		}
		soapResponse := soap.GetSourcesByScanIDResponse{
			GetSourcesByScanIDResult: soap.GetSourcesByScanIDResult{
				CxWSResponseSourcesContent: soap.CxWSResponseSourcesContent{
					CxWSResponseSourceContents: []soap.CxWSResponseSourceContent{
						{Source: fileSources[file1]},
						{Source: fileSources[file3]},
					},
				},
			},
		}
		getSourcesHandler := func(_ string, files []string) (*soap.GetSourcesByScanIDResponse, error) {
			assert.ElementsMatch(t, files, []string{file1, file3})
			return &soapResponse, nil
		}

		ctrl := gomock.NewController(t)
		soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
		soapClientMock.EXPECT().GetSourcesByScanID(scanID, gomock.Any()).DoAndReturn(getSourcesHandler)
		instance := NewRepo(soapClientMock)

		createErr := createFileAndPath(filesToDownload[file2], []byte(fileSources[file2]), metadataFilePerm, metadataFolderPerm)
		assert.NoError(t, createErr)
		err := instance.DownloadSourceFiles(scanID, filesToDownload)
		assert.NoError(t, err)

		assertFileExistWithContent(t, filesToDownload, fileSources)
	})
}

func assertFileExistWithContent(t *testing.T, filesToDownload, fileSources map[string]string) {
	for file, expectedContent := range fileSources {
		_, statErr := os.Stat(filesToDownload[file])
		assert.NoError(t, statErr)
		fileContent, fileContentErr := os.ReadFile(filesToDownload[file])
		assert.NoError(t, fileContentErr)
		assert.Equal(t, expectedContent, string(fileContent))
	}
}
