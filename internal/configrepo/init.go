package configrepo

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/vconf/v2/internal/utils"
	"gopkg.in/src-d/go-git.v4"
	plumbing2 "gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"regexp"
	"sort"
)

var appRe = regexp.MustCompile("refs/(heads|tags)/([a-z|A-Z|0-9|-|.]*)/([a-z|A-Z|0-9|-|.]*)")

type App struct {
	Name     string
	Branches map[string]*plumbing2.Reference
	Versions []*version.Version
}

func NewApp(name string) *App {
	return &App{name, make(map[string]*plumbing2.Reference), make([]*version.Version, 0)}
}

type ConfigRepo struct {
	repo *git.Repository
	Apps map[string]*App
}

func NewConfigRepo(localPath string, cloneOptions *git.CloneOptions) (*ConfigRepo, error) {
	log := logrus.WithField("localPath", localPath)
	log.Info("New Config Repo")
	repo, err := git.PlainOpen(localPath)
	if err == git.ErrRepositoryNotExists {
		log.Warn("no repo found")
		if cloneOptions != nil {
			log.Debugf("cloning it from :%s", cloneOptions.URL)
			repo, err = git.PlainClone(localPath, true, cloneOptions)
		}
	}
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &ConfigRepo{repo, make(map[string]*App)}, nil
}

func (cr *ConfigRepo) Init() error {
	return cr.LoadApps()
}

func (cr *ConfigRepo) addApp(branchRef *plumbing2.Reference) error {
	branchName := branchRef.Name().String()
	appMatches := appRe.FindAllStringSubmatch(branchName, 1)
	if len(appMatches) == 1 && len(appMatches[0]) == 4 {
		appName := appMatches[0][2]
		appStrVersion := appMatches[0][3]
		appVersion, err := version.NewVersion(appStrVersion)
		if err != nil {
			logrus.Warnf("Invalid application version:%s err:%s", appVersion, err)
		} else {
			logrus.Debugf("appName:%s appVersion:%s", appName, appStrVersion)
			if _, appFound := cr.Apps[appName]; !appFound {
				cr.Apps[appName] = NewApp(appName)
			}
			cr.Apps[appName].Branches[appStrVersion] = branchRef
			cr.Apps[appName].Versions = append(cr.Apps[appName].Versions, appVersion)
		}
	} else {
		logrus.Warnf("the branch %s doesn't match with the branch pattern", branchName)
	}
	return nil
}

func (cr *ConfigRepo) LoadApps() error {
	err := cr.loadAppsFromBranches()
	if err != nil {
		logrus.Errorf("Error loading apps from branches:%s", err)
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

func (cr *ConfigRepo) loadAppsFromBranches() error {
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
