package configrepo

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"os"
	"testing"
)

func TestConfigRepo_GetNearestBranch_FullMatch(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	defer os.RemoveAll(localRepo)
	defer os.RemoveAll(remoteRepo)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	branch, err := cfgRepo.GetNearestBranch("app1", "v1.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, branch)
	assert.Contains(t, branch.Name().String(), "app1/v1.0.0")
}

func TestConfigRepo_GetNearestBranch_Between(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	defer os.RemoveAll(localRepo)
	defer os.RemoveAll(remoteRepo)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	branch, err := cfgRepo.GetNearestBranch("app1", "v5.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, branch)
	assert.Contains(t, branch.Name().String(), "refs/tags/app1/v1.0.1")
}

func TestConfigRepo_GetNearestBranch_Over(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	defer os.RemoveAll(localRepo)
	defer os.RemoveAll(remoteRepo)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	branch, err := cfgRepo.GetNearestBranch("app1", "v10.0.0")
	assert.NoError(t, err)
	assert.NotNil(t, branch)
	assert.Contains(t, branch.Name().String(), "app1/v6.0.0")
}

func TestConfigRepo_GetFile(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	defer os.RemoveAll(localRepo)
	defer os.RemoveAll(remoteRepo)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
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
			configContent := getConfigYml(t, cfgRepo, "app1", test.version)
			assert.Equal(t, "dev", configContent["environment"])
			assert.Equal(t, test.expectedVersion, configContent["ver"])
		})
	}
}
