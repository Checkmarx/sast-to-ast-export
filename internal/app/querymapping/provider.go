package querymapping

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/rs/zerolog/log"
)

type RetryableHTTPAdapter interface {
	Get(url string) (*http.Response, error)
}

type (
	Provider struct {
		queryMappings    []QueryMap
		queryMappingPath string
		tmpDir           string
	}
)

const (
	tmpDirPrefix = "query_mapping_"
	fileName     = "mapping.json"
)

func NewProvider(queryMappingPath string, client RetryableHTTPAdapter) (*Provider, error) {
	var mapSource MapSource
	tmpDir := ""
	_, urlErr := url.ParseRequestURI(queryMappingPath)
	if urlErr == nil {
		var tmpFileErr error
		queryMappingPath, tmpDir, tmpFileErr = createTmpFile(queryMappingPath, client)
		if tmpFileErr != nil {
			removeTmpDir(tmpDir)
			return nil, tmpFileErr
		}
	}
	data, err := os.ReadFile(queryMappingPath)
	if err != nil {
		removeTmpDir(tmpDir)
		return nil, err
	}
	if jsonErr := json.Unmarshal(data, &mapSource); jsonErr != nil {
		removeTmpDir(tmpDir)
		return nil, jsonErr
	}
	mapping := mapSource.Mappings

	return &Provider{
		queryMappingPath: queryMappingPath,
		queryMappings:    mapping,
		tmpDir:           tmpDir,
	}, nil
}

func (p *Provider) GetMapping() []QueryMap {
	return p.queryMappings
}

func (p *Provider) GetQueryMappingFilePath() string {
	return p.queryMappingPath
}

func (p *Provider) Clean() error {
	if p.tmpDir != "" {
		return os.RemoveAll(p.tmpDir)
	}

	return nil
}

func createTmpFile(fileUrl string, client RetryableHTTPAdapter) (string, string, error) {
	tmpDir := os.TempDir()
	tmpQueryMappingDir, err := os.MkdirTemp(tmpDir, tmpDirPrefix)
	if err != nil {
		return "", "", err
	}
	tmpFileName := path.Join(tmpQueryMappingDir, fileName)
	out, err := os.Create(tmpFileName)
	if err != nil {
		return "", tmpQueryMappingDir, err
	}
	defer out.Close()
	resp, err := client.Get(fileUrl)
	if err != nil {
		return "", tmpQueryMappingDir, err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", tmpQueryMappingDir, err
	}
	return tmpFileName, tmpQueryMappingDir, nil
}

func removeTmpDir(tmpDir string) {
	if tmpDir == "" {
		return
	}
	delErr := os.RemoveAll(tmpDir)
	if delErr != nil {
		log.Error().Err(delErr).Msgf("Could not remove temporary directory with query mapping file %s", tmpDir)
	}
}
