package configGitRepo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"testing"
	"time"
)

func TestConfigRepo_FetchingEvery(t *testing.T) {
	t.Skip("manual test")
	localRepo, remoteRepo := InitRepos(t)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	assert.NoError(t, cfgRepo.StartFetchingEvery(1*time.Second))
	time.Sleep(3 * time.Second)
	cfgRepo.StopFetching()
	time.Sleep(2 * time.Second)
}

func TestConfigRepo_Pull(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())

	prop3Val := uuid.New().String()
	flContent := fmt.Sprintf("prop3: %s", prop3Val)
	editAndPush(t, remoteRepo, "app1", "v1.0.0","v1.0.0", "config.yml", "added prop3", []byte(flContent))

	configContent := getConfigYml(t, cfgRepo, "app1", "v1.0.0")
	assert.Nil(t, configContent["prop3"])

	assert.NoError(t, cfgRepo.Fetch())
	configContent = getConfigYml(t, cfgRepo, "app1", "v1.0.0")
	assert.Equal(t, prop3Val, configContent["prop3"])
}

func TestConfigRepo_Pull_NewVersion(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())

	prop3Val := uuid.New().String()
	flContent := fmt.Sprintf("prop3: %s", prop3Val)
	editAndPush(t, remoteRepo, "app1", "v1.0.0","v10.0.0", "config.yml", "added prop3", []byte(flContent))

	configContent := getConfigYml(t, cfgRepo, "app1", "v10.0.0")
	assert.Nil(t, configContent["prop3"])

	assert.NoError(t, cfgRepo.Fetch())
	configContent = getConfigYml(t, cfgRepo, "app1", "v10.0.0")
	assert.Equal(t, prop3Val, configContent["prop3"])
}
