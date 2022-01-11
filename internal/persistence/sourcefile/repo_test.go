package sourcefile

import (
	"fmt"
	"os"
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/app/interfaces"

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
		soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
		soapClientMock.EXPECT().GetSourcesByScanID(scanID, gomock.Any()).DoAndReturn(getSourcesHandler)
		instance := NewRepo(soapClientMock)

		err := instance.DownloadSourceFiles(scanID, filesToDownload)
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
		soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
		soapClientMock.EXPECT().GetSourcesByScanID(scanID, gomock.Any()).DoAndReturn(getSourcesHandler)
		instance := NewRepo(soapClientMock)

		createErr := createFileAndPath(filesToDownload[1].LocalName, []byte(fileSources[file2]), metadataFilePerm, metadataFolderPerm)
		assert.NoError(t, createErr)
		err := instance.DownloadSourceFiles(scanID, filesToDownload)
		assert.NoError(t, err)

		assertFileExistWithContent(t, filesToDownload, fileSources)
	})
	t.Run("downloads in multiple batches", func(t *testing.T) {
		tmpDir := t.TempDir()
		var filesToDownload []interfaces.SourceFile
		fileSources := map[string]string{}
		var soapResponses []soap.GetSourcesByScanIDResponse
		for i := 0; i < 25; i++ {
			file := fmt.Sprintf("folder/file%d.kt", i)
			filesToDownload = append(filesToDownload, interfaces.SourceFile{
				RemoteName: file,
				LocalName:  fmt.Sprintf("%s/%s", tmpDir, file),
			})
			fileSources[file] = fmt.Sprintf("file%d", i)

			soapResponseIdx := i / filesPerBatch
			if len(soapResponses) < soapResponseIdx+1 {
				soapResponses = append(soapResponses, soap.GetSourcesByScanIDResponse{
					GetSourcesByScanIDResult: soap.GetSourcesByScanIDResult{
						CxWSResponseSourcesContent: soap.CxWSResponseSourcesContent{
							CxWSResponseSourceContents: []soap.CxWSResponseSourceContent{
								{Source: fileSources[file]},
							},
						},
					},
				})
			} else {
				c := soapResponses[soapResponseIdx].GetSourcesByScanIDResult.CxWSResponseSourcesContent.CxWSResponseSourceContents
				c = append(c, soap.CxWSResponseSourceContent{Source: fileSources[file]})
				soapResponses[soapResponseIdx].GetSourcesByScanIDResult.CxWSResponseSourcesContent.CxWSResponseSourceContents = c
			}
		}

		requestIdx := 0
		getSourcesHandler := func(_ string, _ []string) (*soap.GetSourcesByScanIDResponse, error) {
			r := soapResponses[requestIdx]
			requestIdx++
			return &r, nil
		}

		ctrl := gomock.NewController(t)
		soapClientMock := mock_integration_soap.NewMockAdapter(ctrl)
		soapClientMock.EXPECT().GetSourcesByScanID(scanID, gomock.Any()).DoAndReturn(getSourcesHandler).Times(3)
		instance := NewRepo(soapClientMock)

		err := instance.DownloadSourceFiles(scanID, filesToDownload)
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
