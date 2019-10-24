package configrepo

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
)

func getConfigYml(t *testing.T, cfgRepo *ConfigRepo, appName, targetVersion string) map[string]interface{} {
	cfgFl, err := cfgRepo.GetFile(appName, targetVersion, "config.yml")
	assert.NoError(t, err)
	flReader, err := cfgFl.Reader()
	assert.NoError(t, err)
	defer flReader.Close()
	flCnt, err := ioutil.ReadAll(flReader)
	assert.NoError(t, err)
	configContent := make(map[string]interface{})
	assert.NoError(t, yaml.Unmarshal(flCnt, configContent))
	return configContent
}
