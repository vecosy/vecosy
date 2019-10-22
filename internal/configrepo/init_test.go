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
	v101, err := version.NewVersion("v1.0.1")
	v600, err := version.NewVersion("v6.0.0")
	assert.NoError(t, err)
	assert.Equal(t, cfgRepo.Apps["app1"].Versions[0], v600)
	assert.Equal(t, cfgRepo.Apps["app1"].Versions[1], v101)
	assert.Equal(t, cfgRepo.Apps["app1"].Versions[2], v100)
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

func TestConfigRepo_GetNearestBranch_Between(t *testing.T) {
	cfgRepo, err := NewConfigRepo("../../tests/singleConfigRepo", nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	branch, err := cfgRepo.GetNearestBranch("app1", "v5.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, branch)
	assert.Contains(t, branch.Name().String(), "refs/tags/app1/v1.0.1")
}

func TestConfigRepo_GetNearestBranch_Over(t *testing.T) {
	cfgRepo, err := NewConfigRepo("../../tests/singleConfigRepo", nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	branch, err := cfgRepo.GetNearestBranch("app1", "v10.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, branch)
	assert.Equal(t, branch.Name().String(), "refs/heads/app1/v6.0.0")
}

func TestConfigRepo_GetFile(t *testing.T) {
	cfgRepo, err := NewConfigRepo("../../tests/singleConfigRepo", nil)
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	tests := []struct {
		name            string
		version         string
		expectedVersion string
	}{
		{"version 1.0.0", "1.0.0", "1.0.0"},
		{"version 1.0.1", "1.0.1", "1.0.1"},
		{"version 5.0.0", "5.0.0", "1.0.1"},
		{"version 10.0.0", "10.0.0", "6.0.0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fl, err := cfgRepo.GetFile("app1", test.version, "config.yml")
			assert.NoError(t, err)
			flReader, err := fl.Reader()
			assert.NoError(t, err)
			defer flReader.Close()
			flCnt, err := ioutil.ReadAll(flReader)
			assert.NoError(t, err)
			configContent := make(map[string]interface{})
			assert.NoError(t, yaml.Unmarshal(flCnt, configContent))
			assert.Equal(t, "dev", configContent["environment"])
			assert.Equal(t, test.expectedVersion, configContent["ver"])
		})
	}
}
