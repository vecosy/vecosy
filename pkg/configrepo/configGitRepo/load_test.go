package configGitRepo

import (
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"testing"
)

func TestNewConfigRepo(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	assert.Contains(t, cfgRepo.GetAppsVersions(), "app1")
	v100, err := version.NewVersion("v1.0.0")
	v101, err := version.NewVersion("v1.0.1")
	v600, err := version.NewVersion("v6.0.0")
	assert.NoError(t, err)
	assert.Equal(t, cfgRepo.GetAppsVersions()["app1"][0], v600)
	assert.Equal(t, cfgRepo.GetAppsVersions()["app1"][1], v101)
	assert.Equal(t, cfgRepo.GetAppsVersions()["app1"][2], v100)
}
