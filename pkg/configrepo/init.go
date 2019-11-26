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

func NewApplicationVersion(name, version string) *ApplicationVersion {
	return &ApplicationVersion{
		AppName:    name,
		AppVersion: version,
	}
}

type OnChangeHandler func(changedApplication ApplicationVersion)

type Repo interface {
	Init() error
	GetAppsVersions() map[string][]*version.Version
	GetFile(app *ApplicationVersion, path string) (*RepoFile, error)
	Fetch() error
	GetLastFetch() *time.Time
	StartFetchingEvery(period time.Duration) error
	StopFetching()
	AddOnChangeHandler(handler OnChangeHandler)
}
