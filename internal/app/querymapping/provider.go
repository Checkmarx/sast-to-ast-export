package querymapping

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/checkmarxDev/ast-sast-export/internal/app/common"

	"github.com/rs/zerolog/log"
)

type RetryableHTTPAdapter interface {
	Get(url string) (*http.Response, error)
}

type (
	Provider struct {
		queryMappings []QueryMap
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
	removeTmpDir(tmpDir)
	if err != nil {
		return nil, err
	}
	if jsonErr := json.Unmarshal(data, &mapSource); jsonErr != nil {
		return nil, jsonErr
	}

	return &Provider{
		queryMappings: mapSource.Mappings,
	}, nil
}

func (p *Provider) GetMapping() []QueryMap {
	return p.queryMappings
}

func (p *Provider) AddQueryMapping(language, name, group, sastQueryID string) error {
	for _, mapping := range p.queryMappings {
		if mapping.SastID == sastQueryID {
			return nil
		}
	}

	astID, err := common.GetAstQueryID(language, name, group)
	if err != nil {
		return err
	}

	p.queryMappings = append(p.queryMappings, QueryMap{
		AstID:  astID,
		SastID: sastQueryID,
	})
	return nil
}

func createTmpFile(fileURL string, client RetryableHTTPAdapter) (tmpFileName, tmpQueryMappingDir string, err error) {
	tmpDir := os.TempDir()
	tmpQueryMappingDir, err = os.MkdirTemp(tmpDir, tmpDirPrefix)
	if err != nil {
		return "", "", err
	}
	tmpFileName = path.Join(tmpQueryMappingDir, fileName)
	out, err := os.Create(tmpFileName)
	if err != nil {
		return "", tmpQueryMappingDir, err
	}
	defer out.Close()
	resp, err := client.Get(fileURL)
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
