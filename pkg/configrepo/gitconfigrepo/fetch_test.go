package gitconfigrepo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"gopkg.in/src-d/go-git.v4"
	"testing"
	"time"
)

func TestConfigRepo_FetchingEvery(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	cfgRepo, err := NewGitConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	assert.NoError(t, cfgRepo.StartFetchingEvery(300*time.Millisecond))
	beforeStartFetching := time.Now()
	time.Sleep(500 * time.Millisecond)
	lastFetchTime := &time.Time{}
	*lastFetchTime = *cfgRepo.GetLastFetch()
	t.Logf("lastFetchTime :%+v", lastFetchTime)
	assert.NotNil(t, lastFetchTime)
	assert.True(t, lastFetchTime.After(beforeStartFetching))
	cfgRepo.StopFetching()
	time.Sleep(500 * time.Millisecond)
	assert.True(t, lastFetchTime.Equal(*cfgRepo.GetLastFetch()))
}

func TestConfigRepo_Fetch(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	cfgRepo, err := NewGitConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())

	var changedApp configrepo.ApplicationVersion
	cfgRepo.AddOnChangeHandler(func(changedApplication configrepo.ApplicationVersion) {
		changedApp = changedApplication
	})

	prop3Val := uuid.New().String()
	flContent := fmt.Sprintf("prop3: %s", prop3Val)
	editAndPush(t, remoteRepo, "app1", "v1.0.0", "app1", "v1.0.0", "config.yml", "added prop3", []byte(flContent))

	configContent := getConfigYml(t, cfgRepo, "app1", "v1.0.0")
	assert.Nil(t, configContent["prop3"])

	assert.NoError(t, cfgRepo.Fetch())
	configContent = getConfigYml(t, cfgRepo, "app1", "v1.0.0")
	assert.Equal(t, prop3Val, configContent["prop3"])
	assert.Equal(t, "app1", changedApp.AppName)
	assert.Equal(t, "v1.0.0", changedApp.AppVersion)

}

func TestConfigRepo_Fetch_NewVersion(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	cfgRepo, err := NewGitConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())

	prop3Val := uuid.New().String()
	flContent := fmt.Sprintf("prop3: %s", prop3Val)
	editAndPush(t, remoteRepo, "app1", "v1.0.0", "app1", "v10.0.0", "config.yml", "added prop3", []byte(flContent))

	configContent := getConfigYml(t, cfgRepo, "app1", "v10.0.0")
	assert.Nil(t, configContent["prop3"])

	assert.NoError(t, cfgRepo.Fetch())
	configContent = getConfigYml(t, cfgRepo, "app1", "v10.0.0")
	assert.Equal(t, prop3Val, configContent["prop3"])
}

func TestConfigRepo_Fetch_NewApplication(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	cfgRepo, err := NewGitConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())

	appName := uuid.New().String()
	flContent := fmt.Sprintf("appName: %s", appName)
	editAndPush(t, remoteRepo, "app1", "v1.0.0", "app2", "v2.0.0", "config.yml", "created app2", []byte(flContent))

	assert.NoError(t, cfgRepo.Fetch())
	configContent := getConfigYml(t, cfgRepo, "app2", "v2.0.0")
	assert.Equal(t, appName, configContent["appName"])
}
