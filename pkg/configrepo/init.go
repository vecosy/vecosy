package configrepo

import (
	"github.com/hashicorp/go-version"
	"time"
)

type RepoFile struct {
	Version string
	Content []byte
}
type ApplicationVersion struct {
	AppName    string
	AppVersion string
}

type OnChangeHandler func(changedApplication ApplicationVersion)

type Repo interface {
	Init() error
	GetAppsVersions() map[string][]*version.Version
	GetFile(targetApp, targetVersion, path string) (*RepoFile, error)
	Fetch() error
	StartFetchingEvery(period time.Duration) error
	StopFetching()
	AddOnChangeHandler(handler OnChangeHandler)
}
