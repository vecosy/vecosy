package configGitRepo

import (
	"github.com/hashicorp/go-version"
	"github.com/n3wtron/vconf/v2/pkg/configrepo"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type app struct {
	Name     string
	Branches map[string]*plumbing.Reference
	Versions []*version.Version
}

func newApp(name string) *app {
	return &app{name, make(map[string]*plumbing.Reference), make([]*version.Version, 0)}
}

type ErrorHandlerFn func(err error)

type GitConfigRepo struct {
	repo          *git.Repository
	Apps          map[string]*app
	pullCh        chan bool
	cloneOpts     *git.CloneOptions
	errorsCh      chan error
	errorHandlers []ErrorHandlerFn
}

func NewConfigRepo(localPath string, cloneOpts *git.CloneOptions) (configrepo.Repo, error) {
	log := logrus.WithField("localPath", localPath)
	log.Info("New Config Repo")
	repo, err := git.PlainOpen(localPath)
	if err == git.ErrRepositoryNotExists {
		log.Warn("no repo found")
		if cloneOpts != nil {
			log.Infof("cloning it from :%+v", cloneOpts)
			cloneOpts.Tags = git.AllTags
			cloneOpts.NoCheckout = true
			repo, err = git.PlainClone(localPath, true, cloneOpts)
		}
	}
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &GitConfigRepo{repo, make(map[string]*app), make(chan bool), cloneOpts, make(chan error), make([]ErrorHandlerFn, 0)}, nil
}

func (cr *GitConfigRepo) Init() error {
	cr.errorHandlerManager()
	return cr.loadApps()
}

func (cr *GitConfigRepo) GetAppsVersions() map[string][]*version.Version {
	result := make(map[string][]*version.Version)
	for app, branch := range cr.Apps {
		result[app] = branch.Versions
	}
	return result
}
