package sourcefile

import (
	"fmt"
	"os"
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"

	"github.com/checkmarxDev/ast-sast-export/internal/integration/soap"
	mock_integration_soap "github.com/checkmarxDev/ast-sast-export/test/mocks/integration/soap"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRepo_DownloadSourceFiles(t *testing.T) {
	scanID := "1000000"
	file1 := "project/folder1/file1.go"
	file2 := "project/folder1/file2.go"
	file3 := "project/folder2/file3.go"

	t.Run("success case", func(t *testing.T) {
		tmpDir := t.TempDir()
		filesToDownload := []interfaces.SourceFile{
			{RemoteName: file1, LocalName: fmt.Sprintf("%s/%s", tmpDir, file1)},
			{RemoteName: file2, LocalName: fmt.Sprintf("%s/%s", tmpDir, file2)},
			{RemoteName: file3, LocalName: fmt.Sprintf("%s/%s", tmpDir, file3)},
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
		defer ctrl.Finish()
		soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
		soapClientMock.EXPECT().GetSourcesByScanID(scanID, gomock.Any()).DoAndReturn(getSourcesHandler)

		instance := NewRepo(soapClientMock)
		err := instance.DownloadSourceFiles(scanID, filesToDownload, "")
		assert.NoError(t, err)

		assertFileExistWithContent(t, filesToDownload, fileSources)
	})

	t.Run("only downloads files that don't exist in local filesystem", func(t *testing.T) {
		tmpDir := t.TempDir()
		filesToDownload := []interfaces.SourceFile{
			{RemoteName: file1, LocalName: fmt.Sprintf("%s/%s", tmpDir, file1)},
			{RemoteName: file2, LocalName: fmt.Sprintf("%s/%s", tmpDir, file2)},
			{RemoteName: file3, LocalName: fmt.Sprintf("%s/%s", tmpDir, file3)},
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
		defer ctrl.Finish()
		soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
		soapClientMock.EXPECT().GetSourcesByScanID(scanID, gomock.Any()).DoAndReturn(getSourcesHandler)

		instance := NewRepo(soapClientMock)

		createErr := createFileAndPath(filesToDownload[1].LocalName, []byte(fileSources[file2]), metadataFilePerm, metadataFolderPerm)
		assert.NoError(t, createErr)

		err := instance.DownloadSourceFiles(scanID, filesToDownload, "")
		assert.NoError(t, err)

		assertFileExistWithContent(t, filesToDownload, fileSources)
	})

}

func assertFileExistWithContent(t *testing.T, filesToDownload []interfaces.SourceFile, fileSources map[string]string) {
	for _, sourceFile := range filesToDownload {
		_, statErr := os.Stat(sourceFile.LocalName)
		assert.NoError(t, statErr)
		fileContent, fileContentErr := os.ReadFile(sourceFile.LocalName)
		assert.NoError(t, fileContentErr)
		assert.Equal(t, fileSources[sourceFile.RemoteName], string(fileContent))
	}
}

func TestRepo_DownloadSourceFiles_WithExclusions(t *testing.T) {
	scanID := "1000001"
	tmpDir := t.TempDir()

	// Create an exclude file with paths to exclude
	excludeFilePath := fmt.Sprintf("%s/exclude.txt", tmpDir)
	excludedFile := "project/folder1/file1.go"
	err := os.WriteFile(excludeFilePath, []byte(excludedFile+"\n"), 0644)
	assert.NoError(t, err)

	filesToDownload := []interfaces.SourceFile{
		{RemoteName: excludedFile, LocalName: fmt.Sprintf("%s/%s", tmpDir, excludedFile)},
		{RemoteName: "project/folder1/file2.go", LocalName: fmt.Sprintf("%s/project/folder1/file2.go", tmpDir)},
	}

	soapResponse := soap.GetSourcesByScanIDResponse{
		GetSourcesByScanIDResult: soap.GetSourcesByScanIDResult{
			CxWSResponseSourcesContent: soap.CxWSResponseSourcesContent{
				CxWSResponseSourceContents: []soap.CxWSResponseSourceContent{
					{Source: "file2 content"},
				},
			},
		},
	}
	getSourcesHandler := func(_ string, files []string) (*soap.GetSourcesByScanIDResponse, error) {
		assert.ElementsMatch(t, files, []string{"project/folder1/file2.go"}) // Ensure excluded file is not requested
		return &soapResponse, nil
	}

	ctrl := gomock.NewController(t)
	soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
	soapClientMock.EXPECT().GetSourcesByScanID(scanID, gomock.Any()).DoAndReturn(getSourcesHandler)
	instance := NewRepo(soapClientMock)

	err = instance.DownloadSourceFiles(scanID, filesToDownload, excludeFilePath)
	assert.NoError(t, err)

	// Check that excluded file was NOT created
	_, statErr := os.Stat(filesToDownload[0].LocalName)
	assert.Error(t, statErr, "excluded file should not exist")

	// Check that non-excluded file was created
	fileContent, fileErr := os.ReadFile(filesToDownload[1].LocalName)
	assert.NoError(t, fileErr)
	assert.Equal(t, "file2 content", string(fileContent))
}
