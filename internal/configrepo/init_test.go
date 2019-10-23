package configrepo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var editorSignature = &object.Signature{
	Name:  "Config Editor",
	Email: "editor@cfg.local",
	When:  time.Now(),
}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	m.Run()
}

func InitRepos(t *testing.T) (string, string) {
	localTmpRepo := fmt.Sprintf("%s/%s", os.TempDir(), uuid.New().String())
	remoteTmpRepo := fmt.Sprintf("%s/%s", os.TempDir(), uuid.New().String())
	assert.NoError(t, archiver.Unarchive("../../tests/singleConfigRepo.tgz", remoteTmpRepo))
	return localTmpRepo, remoteTmpRepo + "/singleConfigRepo"
}

func TestNewConfigRepo(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	defer os.RemoveAll(localRepo)
	defer os.RemoveAll(remoteRepo)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
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

func TestConfigRepo_PullingEvery(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	defer os.RemoveAll(localRepo)
	defer os.RemoveAll(remoteRepo)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())
	assert.NoError(t, cfgRepo.StartPullingEvery(1*time.Second))
	time.Sleep(3 * time.Second)
	cfgRepo.StopPulling()
	time.Sleep(2 * time.Second)
}

func TestConfigRepo_Pull(t *testing.T) {
	localRepo, remoteRepo := InitRepos(t)
	defer os.RemoveAll(localRepo)
	defer os.RemoveAll(remoteRepo)
	cfgRepo, err := NewConfigRepo(localRepo, &git.CloneOptions{URL: remoteRepo})
	assert.NoError(t, err)
	assert.NotNil(t, cfgRepo)
	assert.NoError(t, cfgRepo.Init())

	editorRepoPath := fmt.Sprintf("%s/%s", os.TempDir(), uuid.New().String())
	t.Logf("editorPath:%s",editorRepoPath)
	//defer os.RemoveAll(editorRepoPath)
	editorRepo, err := git.PlainClone(editorRepoPath, false, &git.CloneOptions{URL: remoteRepo, NoCheckout: true})
	assert.NoError(t, err)

	wk, err := editorRepo.Worktree()
	assert.NoError(t, err)
	err = wk.Checkout(&git.CheckoutOptions{Branch: "refs/remotes/origin/app1/v1.0.0",})
	assert.NoError(t, err)
	fl, err := wk.Filesystem.OpenFile("config.yml", os.O_RDWR, 0644)
	assert.NoError(t, err)
	defer fl.Close()
	prop3Val := uuid.New().String()
	_, err = fl.Write([]byte(fmt.Sprintf("prop3: %s", prop3Val)))
	assert.NoError(t, err)
	assert.NoError(t, fl.Close())
	_, err = wk.Add("config.yml")
	assert.NoError(t, err)
	_, err = wk.Commit("added prop3", &git.CommitOptions{All: true, Author: editorSignature,})
	assert.NoError(t, err)
	assert.NoError(t, editorRepo.Push(&git.PushOptions{}))

	configContent := getConfigYml(t, cfgRepo, "app1", "v1.0.0")
	assert.Nil(t, configContent["prop3"])

	assert.NoError(t, cfgRepo.Pull())
	configContent = getConfigYml(t, cfgRepo, "app1", "v1.0.0")
	assert.Equal(t, prop3Val, configContent["prop3"])

}

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
