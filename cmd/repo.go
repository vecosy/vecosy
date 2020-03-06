package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"github.com/vecosy/vecosy/v2/pkg/configrepo/gitconfigrepo"
	ssh2 "golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"path"
	"path/filepath"
	"strings"
)

func initRepo() configrepo.Repo {
	repoURL := viper.GetString("repo.remote.url")
	if strings.Contains(repoURL, "file://") {
		repoPath := strings.Replace(repoURL, "file://", "", 1)
		if !path.IsAbs(repoPath) {
			repoURL, _ = filepath.Abs(repoPath)
		}
	}
	auth, err := getAuth()
	if err != nil {
		logrus.Fatalf("error initializing repo auth:%s", err)
	}
	cfgRepo, err := gitconfigrepo.NewGitConfigRepo(viper.GetString("repo.local.path"), &git.CloneOptions{URL: repoURL, Auth: auth})
	if err != nil {
		logrus.Fatalf("error initializing repo:%s", err)
	}
	err = cfgRepo.Init()
	if err != nil {
		logrus.Fatalf("error loading the config repo:%s", err)
	}
	pullEvery := viper.GetDuration("repo.remote.pullEvery")
	logrus.Infof("Fetch repo every :%s", pullEvery)
	err = cfgRepo.StartFetchingEvery(pullEvery)
	if err != nil {
		logrus.Fatalf("error fetching the repo:%s", err)
	}
	return cfgRepo
}

func getAuth() (transport.AuthMethod, error) {
	authType := viper.GetString("repo.remote.auth.type")
	username := viper.GetString("repo.remote.auth.username")
	switch authType {
	case "pubKey":
		keyFile := viper.GetString("repo.remote.auth.keyFile")
		keyFilePassword := viper.GetString("repo.remote.auth.keyFilePassword")
		sshAuth, err := ssh.NewPublicKeysFromFile(username, keyFile, keyFilePassword)
		if err != nil {
			return nil, err
		}
		sshAuth.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
		return sshAuth, nil
	case "plain":
		password := viper.GetString("repo.remote.auth.password")
		return &ssh.Password{
			User:                  username,
			Password:              password,
			HostKeyCallbackHelper: ssh.HostKeyCallbackHelper{},
		}, nil
	case "http":
		password := viper.GetString("repo.remote.auth.password")
		return &http.BasicAuth{
			Username: username,
			Password: password,
		}, nil
	default:
		return nil, nil
	}
}
