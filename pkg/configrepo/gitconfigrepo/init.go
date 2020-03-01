package gitconfigrepo

import (
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"sync"
	"time"
)

type app struct {
	Name     string
	Branches map[string]*plumbing.Reference
	Versions []*version.Version
}

func newApp(name string) *app {
	return &app{name, make(map[string]*plumbing.Reference), make([]*version.Version, 0)}
}

// ErrorHandlerFn represent an error handler function
type ErrorHandlerFn func(err error)

// GitConfigRepo represent a git configuration repository
type GitConfigRepo struct {
	repo            *git.Repository
	Apps            map[string]*app
	fetchCh         chan bool
	lastFetch       *time.Time
	lastFetchMutex  sync.Mutex
	cloneOpts       *git.CloneOptions
	errorsCh        chan error
	errorHandlers   []ErrorHandlerFn
	changesHandlers []configrepo.OnChangeHandler
}

// NewGitConfigRepo instantiate a new GIT configuration repository
func NewGitConfigRepo(localPath string, cloneOpts *git.CloneOptions) (configrepo.Repo, error) {
	log := logrus.WithField("localPath", localPath)
	log.Info("New Config Repo")
	repo, err := git.PlainOpen(localPath)
	if err == git.ErrRepositoryNotExists {
		log.Warn("no repo found")
		if cloneOpts != nil {
			cloneOpts.Tags = git.AllTags
			cloneOpts.NoCheckout = true
			log.Infof("cloning it from :%+v", cloneOpts)
			repo, err = git.PlainClone(localPath, false, cloneOpts)
		}
	}
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &GitConfigRepo{
		repo:            repo,
		Apps:            make(map[string]*app),
		fetchCh:         make(chan bool),
		lastFetch:       nil,
		cloneOpts:       cloneOpts,
		errorsCh:        make(chan error),
		errorHandlers:   make([]ErrorHandlerFn, 0),
		changesHandlers: make([]configrepo.OnChangeHandler, 0),
	}, nil
}

// Init initialize the git repository
func (cr *GitConfigRepo) Init() (err error) {
	cr.errorHandlerManager()
	cr.Apps, err = cr.loadApps()
	return
}

// GetAppsVersions returns a appName-> list of version
func (cr *GitConfigRepo) GetAppsVersions() map[string][]*version.Version {
	result := make(map[string][]*version.Version)
	for app, branch := range cr.Apps {
		result[app] = branch.Versions
	}
	return result
}

// AddOnChangeHandler add a new change handler to the git repo
func (cr *GitConfigRepo) AddOnChangeHandler(handler configrepo.OnChangeHandler) {
	cr.changesHandlers = append(cr.changesHandlers, handler)
}
