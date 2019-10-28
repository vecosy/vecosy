package configGitRepo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/n3wtron/vconf/v2/pkg/configrepo"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/yaml.v2"
	"os"
	"testing"
)

func getConfigYml(t *testing.T, cfgRepo configrepo.Repo, appName, targetVersion string) map[string]interface{} {
	cfgFl, err := cfgRepo.GetFile(appName, targetVersion, "config.yml")
	assert.NoError(t, err)
	configContent := make(map[string]interface{})
	assert.NoError(t, yaml.Unmarshal(cfgFl.Content, configContent))
	return configContent
}

func editAndPush(t *testing.T, remoteRepo, app, srcVersion, dstVersion, fileName, commitMsg string, content []byte) {
	editorRepoPath := fmt.Sprintf("%s/%s", testBasicPath, uuid.New().String())
	t.Logf("editorPath:%s", editorRepoPath)
	editorRepo, err := git.PlainClone(editorRepoPath, false, &git.CloneOptions{URL: remoteRepo, NoCheckout: true})
	assert.NoError(t, err)

	wk, err := editorRepo.Worktree()
	assert.NoError(t, err)
	remoteAppBranch := plumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s/%s", app, srcVersion))
	err = wk.Checkout(&git.CheckoutOptions{Branch: remoteAppBranch, Force: true})
	assert.NoError(t, err)

	fl, err := wk.Filesystem.OpenFile(fileName, os.O_RDWR, 0644)
	assert.NoError(t, err)
	defer fl.Close()
	_, err = fl.Write(content)
	assert.NoError(t, err)
	assert.NoError(t, fl.Close())
	_, err = wk.Add(fileName)
	assert.NoError(t, err)
	_, err = wk.Commit(commitMsg, &git.CommitOptions{All: true, Author: editorSignature,})
	assert.NoError(t, err)

	head, err := editorRepo.Head()
	assert.NoError(t, err)
	localBranch := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s/%s", app, dstVersion))
	branchRef := plumbing.NewHashReference(localBranch, head.Hash())
	assert.NoError(t, editorRepo.Storer.SetReference(branchRef))

	reference := config.RefSpec(fmt.Sprintf("%s:%s", localBranch, localBranch))
	assert.NoError(t, editorRepo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{reference},
	}))
}
