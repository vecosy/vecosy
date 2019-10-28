package configrepo

import (
	"github.com/hashicorp/go-version"
	"time"
)

type RepoFile struct {
	Version string
	Content []byte
}

type Repo interface {
	Init() error
	GetAppsVersions() map[string][]*version.Version
	GetFile(targetApp, targetVersion, path string) (*RepoFile, error)
	Pull() error
	StartPullingEvery(period time.Duration) error
	StopPulling()
}
