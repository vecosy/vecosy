package configGitRepo

import (
	"github.com/hashicorp/go-version"
	"github.com/n3wtron/vconf/v2/internal/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"regexp"
	"sort"
)
var appRe = regexp.MustCompile(".*/([a-z|A-Z|0-9|-|.]*)/([a-z|A-Z|0-9|-|.]*)")

func (cr *GitConfigRepo) addApp(branchRef *plumbing.Reference) error {
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
				cr.Apps[appName] = newApp(appName)
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

func (cr *GitConfigRepo) loadApps() error {
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

	return storer.NewReferenceFilteredIter(func(ref *plumbing.Reference) bool {
		return ref.Name().IsRemote()
	}, refs), nil
}

func (cr *GitConfigRepo) loadAppsFromRemoteBranches() error {
	branches, err := remoteBranches(cr.repo.Storer)
	if err != nil {
		return err
	}
	return branches.ForEach(cr.addApp)
}

func (cr *GitConfigRepo) loadAppsFromLocalBranches() error {
	branches, err := cr.repo.Branches()
	if err != nil {
		return err
	}
	return branches.ForEach(cr.addApp)
}

func (cr *GitConfigRepo) loadAppsFromTags() error {
	tags, err := cr.repo.Tags()
	if err != nil {
		return err
	}
	return tags.ForEach(cr.addApp)
}
