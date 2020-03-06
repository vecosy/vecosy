package configrepo

import (
	"github.com/hashicorp/go-version"
	"time"
)

// RepoFile represent a repository file
type RepoFile struct {
	Version string
	Content []byte
}

// ApplicationVersion represent the couple name+version
type ApplicationVersion struct {
	AppName    string
	AppVersion string
}

// NewApplicationVersion create a new ApplicationVersion (name+version) instance
func NewApplicationVersion(name, version string) *ApplicationVersion {
	return &ApplicationVersion{
		AppName:    name,
		AppVersion: version,
	}
}

// OnChangeHandler function handler executed on every repo changes
type OnChangeHandler func(changedApplication ApplicationVersion)

// Repo represent a config repository
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
