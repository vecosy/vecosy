package gitconfigrepo

import (
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"regexp"
	"sort"
)

var appRe = regexp.MustCompile(`.*/([a-z|A-Z|0-9|\-|.]*)/([a-z|A-Z|0-9|\-|.]*)`)

func addApp(branchRef *plumbing.Reference, apps map[string]*app) error {
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
			if _, appFound := apps[appName]; !appFound {
				apps[appName] = newApp(appName)
			}
			if _, alreadyPresent := apps[appName].Branches[appStrVersion]; !alreadyPresent {
				apps[appName].Versions = append(apps[appName].Versions, appVersion)
			}
			apps[appName].Branches[appStrVersion] = branchRef
		}
	} else {
		logrus.Warnf("the branch %s doesn't match with the branch pattern", branchName)
	}
	return nil
}

func (cr *GitConfigRepo) loadApps() (map[string]*app, error) {
	newApps := make(map[string]*app)
	err := cr.loadAppsFromRemoteBranches(newApps)
	if err != nil {
		logrus.Errorf("Error loading apps from remote branches:%s", err)
		return nil, err
	}

	err = cr.loadAppsFromLocalBranches(newApps)
	if err != nil {
		logrus.Errorf("Error loading apps from local branches:%s", err)
		return nil, err
	}

	err = cr.loadAppsFromTags(newApps)
	if err != nil {
		logrus.Errorf("Error loading apps from tags:%s", err)
		return nil, err
	}

	for appName, app := range newApps {
		logrus.Debugf("sorting app:%s versions", appName)
		sort.Sort(version.Collection(app.Versions))
		utils.ReverseVersion(app.Versions)
		logrus.Infof("app:%s Sorted Versions:%+v", appName, app.Versions)
	}
	return newApps, nil
}

func remoteBranches(s storer.ReferenceStorer) (storer.ReferenceIter, error) {
	refs, err := s.IterReferences()
	if err != nil {
		return nil, err
	}

	return storer.NewReferenceFilteredIter(func(ref *plumbing.Reference) bool {
		return ref.Name().IsRemote()
	}, refs), nil
}

func (cr *GitConfigRepo) loadAppsFromRemoteBranches(apps map[string]*app) error {
	branches, err := remoteBranches(cr.repo.Storer)
	if err != nil {
		return err
	}
	return branches.ForEach(func(reference *plumbing.Reference) error {
		return addApp(reference, apps)
	})
}

func (cr *GitConfigRepo) loadAppsFromLocalBranches(apps map[string]*app) error {
	branches, err := cr.repo.Branches()
	if err != nil {
		return err
	}
	return branches.ForEach(func(reference *plumbing.Reference) error {
		return addApp(reference, apps)
	})
}

func (cr *GitConfigRepo) loadAppsFromTags(apps map[string]*app) error {
	tags, err := cr.repo.Tags()
	if err != nil {
		return err
	}
	return tags.ForEach(func(reference *plumbing.Reference) error {
		return addApp(reference, apps)
	})
}
