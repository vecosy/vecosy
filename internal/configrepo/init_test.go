package configrepo

import (
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	m.Run()
}

func TestNewConfigRepo(t *testing.T) {
	cfgRepo, err := NewConfigRepo("../../tests/singleConfigRepo", nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	assert.Contains(t, cfgRepo.Apps, "app1")
	v100, err := version.NewVersion("v1.0.0")
	assert.NoError(t, err)
	v101, err := version.NewVersion("v1.0.1")
	assert.NoError(t, err)
	assert.Equal(t, cfgRepo.Apps["app1"].Versions[0], v100)
	assert.Equal(t, cfgRepo.Apps["app1"].Versions[1], v101)
}

func TestConfigRepo_GetNearestBranch_FullMatch(t *testing.T) {
	cfgRepo, err := NewConfigRepo("../../tests/singleConfigRepo", nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	branch, err := cfgRepo.GetNearestBranch("app1", "v1.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, branch)
	assert.Equal(t, branch.Name().String(), "refs/heads/app1/v1.0.0")
}

func TestConfigRepo_GetNearestBranch_Over(t *testing.T) {
	cfgRepo, err := NewConfigRepo("../../tests/singleConfigRepo", nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	branch, err := cfgRepo.GetNearestBranch("app1", "v5.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, branch)
	assert.Equal(t, branch.Name().String(), "refs/heads/app1/v1.0.1")
}

func TestConfigRepo_GetFile(t *testing.T) {
	cfgRepo, err := NewConfigRepo("../../tests/singleConfigRepo", nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	fl, err := cfgRepo.GetFile("app1", "1.0.0", "dev", "config.yml")
	assert.NoError(t, err)
	flReader, err := fl.Reader()
	assert.NoError(t, err)
	defer flReader.Close()
	flCnt, err := ioutil.ReadAll(flReader)
	assert.NoError(t, err)
	configContent := make(map[string]interface{})
	assert.NoError(t, yaml.Unmarshal(flCnt, configContent))
	assert.Equal(t, "dev", configContent["environment"])

}
