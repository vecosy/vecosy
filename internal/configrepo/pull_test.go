package configrepo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
	"testing"
	"time"
)

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

	editorRepoPath := fmt.Sprintf("%s/vconf/%s", os.TempDir(), uuid.New().String())
	t.Logf("editorPath:%s", editorRepoPath)
	//defer os.RemoveAll(editorRepoPath)
	editorRepo, err := git.PlainClone(editorRepoPath, false, &git.CloneOptions{URL: remoteRepo, NoCheckout: true})
	assert.NoError(t, err)

	wk, err := editorRepo.Worktree()
	assert.NoError(t, err)
	err = wk.Checkout(&git.CheckoutOptions{Branch: "refs/remotes/origin/app1/v1.0.0", Force: true})
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

	head, err := editorRepo.Head()
	assert.NoError(t, err)
	branchRef := plumbing.NewHashReference("refs/heads/app1/v1.0.0", head.Hash())
	assert.NoError(t, editorRepo.Storer.SetReference(branchRef))

	assert.NoError(t, editorRepo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/app1/v1.0.0:refs/heads/app1/v1.0.0"},
	}))

	configContent := getConfigYml(t, cfgRepo, "app1", "v1.0.0")
	assert.Nil(t, configContent["prop3"])

	assert.NoError(t, cfgRepo.Pull())
	configContent = getConfigYml(t, cfgRepo, "app1", "v1.0.0")
	assert.Equal(t, prop3Val, configContent["prop3"])
}
