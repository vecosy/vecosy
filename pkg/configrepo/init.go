package configrepo

import (
	"github.com/hashicorp/go-version"
	"time"
)

type Repo interface {
	Init() error
	GetAppsVersions() map[string][]*version.Version
	GetFile(targetApp, targetVersion, path string) ([]byte, error)
	Pull() error
	StartPullingEvery(period time.Duration) error
	StopPulling()
}
