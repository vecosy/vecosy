package configrepo

import (
	"github.com/hashicorp/go-version"
	"time"
)

type RepoFile struct {
	Version string
	Content []byte
}
type OnChangeHandler func(appName, appVersion string)

type Repo interface {
	Init() error
	GetAppsVersions() map[string][]*version.Version
	GetFile(targetApp, targetVersion, path string) (*RepoFile, error)
	Pull() error
	StartPullingEvery(period time.Duration) error
	StopPulling()
	AddOnChangeHandler(handler OnChangeHandler)
}
