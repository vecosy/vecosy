package configrepo

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/vconf/v2/internal/utils"
	"gopkg.in/src-d/go-git.v4"
	plumbing2 "gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"regexp"
	"sort"
	"time"
)

var appRe = regexp.MustCompile(".*/([a-z|A-Z|0-9|-|.]*)/([a-z|A-Z|0-9|-|.]*)")

type App struct {
	Name     string
	Branches map[string]*plumbing2.Reference
	Versions []*version.Version
}

func NewApp(name string) *App {
	return &App{name, make(map[string]*plumbing2.Reference), make([]*version.Version, 0)}
}

type ErrorHandlerFn func(err error)

type ConfigRepo struct {
	repo          *git.Repository
	Apps          map[string]*App
	pullCh        chan bool
	cloneOpts     *git.CloneOptions
	errorsCh      chan error
	errorHandlers []ErrorHandlerFn
}

func NewConfigRepo(localPath string, cloneOpts *git.CloneOptions) (*ConfigRepo, error) {
	log := logrus.WithField("localPath", localPath)
	log.Info("New Config Repo")
	repo, err := git.PlainOpen(localPath)
	if err == git.ErrRepositoryNotExists {
		log.Warn("no repo found")
		if cloneOpts != nil {
			log.Debugf("cloning it from :%+v", cloneOpts)
			cloneOpts.Tags = git.AllTags
			cloneOpts.NoCheckout = true
			repo, err = git.PlainClone(localPath, true, cloneOpts)
		}
	}
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &ConfigRepo{repo, make(map[string]*App), make(chan bool), cloneOpts, make(chan error), make([]ErrorHandlerFn, 0)}, nil
}

func (cr *ConfigRepo) Init() error {
	cr.errorHandlerManager()
	return cr.LoadApps()
}

func (cr *ConfigRepo) addApp(branchRef *plumbing2.Reference) error {
	logrus.Debugf("analyzing reference :%s", branchRef.Name())
	branchName := branchRef.Name().String()
	appMatches := appRe.FindAllStringSubmatch(branchName, 1)
	if len(appMatches) == 1 && len(appMatches[0]) == 3 {
		appName := appMatches[0][1]
		appStrVersion := appMatches[0][2]
		appVersion, err := version.NewVersion(appStrVersion)
		if err != nil {
			logrus.Warnf("Invalid application version:%s err:%s", appVersion, err)
		} else {
			logrus.Debugf("appName:%s appVersion:%s", appName, appStrVersion)
			if _, appFound := cr.Apps[appName]; !appFound {
				cr.Apps[appName] = NewApp(appName)
			}
			if _, alreadyPresent := cr.Apps[appName].Branches[appStrVersion]; !alreadyPresent {
				cr.Apps[appName].Versions = append(cr.Apps[appName].Versions, appVersion)
			}
			cr.Apps[appName].Branches[appStrVersion] = branchRef
		}
	} else {
		logrus.Warnf("the branch %s doesn't match with the branch pattern", branchName)
	}
	return nil
}

func (cr *ConfigRepo) LoadApps() error {
	err := cr.loadAppsFromRemoteBranches()
	if err != nil {
		logrus.Errorf("Error loading apps from remote branches:%s", err)
		return err
	}

	err = cr.loadAppsFromLocalBranches()
	if err != nil {
		logrus.Errorf("Error loading apps from local branches:%s", err)
		return err
	}

	err = cr.loadAppsFromTags()
	if err != nil {
		logrus.Errorf("Error loading apps from tags:%s", err)
		return err
	}

	for appName, app := range cr.Apps {
		logrus.Debugf("sorting app:%s versions", appName)
		sort.Sort(version.Collection(app.Versions))
		utils.ReverseVersion(app.Versions)
		logrus.Infof("app:%s Sorted Versions:%+v", appName, app.Versions)
	}
	return nil
}

func remoteBranches(s storer.ReferenceStorer) (storer.ReferenceIter, error) {
	refs, err := s.IterReferences()
	if err != nil {
		return nil, err
	}

	return storer.NewReferenceFilteredIter(func(ref *plumbing2.Reference) bool {
		return ref.Name().IsRemote()
	}, refs), nil
}

func (cr *ConfigRepo) loadAppsFromRemoteBranches() error {
	branches, err := remoteBranches(cr.repo.Storer)
	if err != nil {
		return err
	}
	return branches.ForEach(cr.addApp)
}

func (cr *ConfigRepo) loadAppsFromLocalBranches() error {
	branches, err := cr.repo.Branches()
	if err != nil {
		return err
	}
	return branches.ForEach(cr.addApp)
}

func (cr *ConfigRepo) loadAppsFromTags() error {
	tags, err := cr.repo.Tags()
	if err != nil {
		return err
	}
	return tags.ForEach(cr.addApp)
}

func (cr *ConfigRepo) GetNearestBranch(targetApp, targetVersion string) (*plumbing2.Reference, error) {
	app, appFound := cr.Apps[targetApp]
	if !appFound {
		return nil, fmt.Errorf("no app found with name %s", targetApp)
	}
	constraint, err := version.NewConstraint(fmt.Sprintf("<=%s", targetVersion))
	if err != nil {
		return nil, err
	}
	for _, chkVer := range app.Versions {
		if constraint.Check(chkVer) {
			return app.Branches[chkVer.Original()], nil
		}
	}
	return nil, fmt.Errorf("no branch found for target chkVer:%s", targetVersion)
}

func (cr *ConfigRepo) GetFile(targetApp, targetVersion, path string) (*object.File, error) {
	branchRef, err := cr.GetNearestBranch(targetApp, targetVersion)
	if err != nil {
		return nil, err
	}
	commit, err := cr.repo.CommitObject(branchRef.Hash())
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	return tree.File(path)
}

func (cr *ConfigRepo) StartPullingEvery(period time.Duration) error {
	t := time.NewTicker(period)
	go func() {
		for {
			select {
			case t := <-t.C:
				logrus.Debugf("Auto pull :%+s", t)
				cr.pushError(cr.Pull())
			case <-cr.pullCh:
				t.Stop()
				return
			}
		}
	}()
	return nil
}

func (cr *ConfigRepo) StopPulling() {
	cr.pullCh <- true
}

func (cr *ConfigRepo) Pull() error {
	logrus.Info("Pull")
	if cr.cloneOpts != nil {
		fetchOpts := &git.FetchOptions{Auth: cr.cloneOpts.Auth, Force: true, Tags: git.AllTags}
		err := cr.repo.Fetch(fetchOpts)
		if err != nil {
			if err != git.NoErrAlreadyUpToDate {
				logrus.Errorf("Error pulling :%s", err)
				return err
			} else {
				logrus.Info("already up to date")
			}
		}
		return cr.LoadApps()
	} else {
		logrus.Warn("Cannot pull:no remote information found")
	}
	return nil
}

func (cr *ConfigRepo) pushError(err error) {
	if err != nil {
		cr.errorsCh <- err
	}
}

func (cr *ConfigRepo) addErrorListener(fn ErrorHandlerFn) {
	cr.errorHandlers = append(cr.errorHandlers, fn)
}

func (cr *ConfigRepo) errorHandlerManager() {
	go func() {
		for {
			select {
			case err := <-cr.errorsCh:
				for _, errFn := range cr.errorHandlers {
					errFn(err)
				}
			}
		}
	}()
}
