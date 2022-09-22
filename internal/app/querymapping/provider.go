package querymapping

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

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

func NewProvider(queryMappingPath string) (*Provider, error) {
	var mapSource MapSource
	tmpDir := ""
	_, urlErr := url.ParseRequestURI(queryMappingPath)
	if urlErr == nil {
		var tmpFileErr error
		queryMappingPath, tmpDir, tmpFileErr = createTmpFile(queryMappingPath)
		if tmpFileErr != nil {
			return nil, tmpFileErr
		}
	}
	data, err := os.ReadFile(queryMappingPath)
	if err != nil {
		return nil, err
	}
	if jsonErr := json.Unmarshal(data, &mapSource); jsonErr != nil {
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

func createTmpFile(fileUrl string) (string, string, error) {
	tmpDir := os.TempDir()
	tmpQueryMappingDir, err := os.MkdirTemp(tmpDir, tmpDirPrefix)
	if err != nil {
		return "", "", err
	}
	tmpFileName := path.Join(tmpQueryMappingDir, fileName)
	out, err := os.Create(tmpFileName)
	if err != nil {
		return "", "", err
	}
	defer out.Close()
	resp, err := http.Get(fileUrl)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", "", err
	}
	return tmpFileName, tmpQueryMappingDir, nil
}
