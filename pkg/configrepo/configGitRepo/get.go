package configGitRepo

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io/ioutil"
)

func (cr *GitConfigRepo) GetNearestBranch(targetApp *configrepo.ApplicationVersion) (*plumbing.Reference, error) {
	app, appFound := cr.Apps[targetApp.AppName]
	if !appFound {
		return nil, configrepo.ApplicationNotFoundError
	}
	constraint, err := version.NewConstraint(fmt.Sprintf("<=%s", targetApp.AppVersion))
	if err != nil {
		return nil, err
	}
	for _, chkVer := range app.Versions {
		if constraint.Check(chkVer) {
			return app.Branches[chkVer.Original()], nil
		}
	}
	return nil, fmt.Errorf("no branch found for target chkVer:%s", targetApp.AppVersion)
}

func (cr *GitConfigRepo) GetFile(targetApp *configrepo.ApplicationVersion, path string) (*configrepo.RepoFile, error) {
	log := logrus.WithField("method", "GetFile").WithField("targetApp", targetApp).WithField("path", path)
	branchRef, err := cr.GetNearestBranch(targetApp)
	if err != nil {
		return nil, err
	}
	commit, err := cr.repo.CommitObject(branchRef.Hash())
	if err != nil {
		log.Errorf("Error getting the commit object:%s", err)
		return nil, err
	}
	log.Debugf("found commit: %s", commit.Hash.String())
	tree, err := commit.Tree()
	if err != nil {
		log.Errorf("Error getting the tree:%s", err)
		return nil, err
	}

	fl, err := tree.File(path)
	if err != nil {
		log.Errorf("Error getting the file:%s", err)
		return nil, err
	}
	flReader, err := fl.Reader()
	if err != nil {
		log.Errorf("Error creating the file reader:%s", err)
		return nil, err
	}
	defer flReader.Close()
	result := &configrepo.RepoFile{Version: commit.Hash.String()}
	result.Content, err = ioutil.ReadAll(flReader)
	if err != nil {
		log.Errorf("Error reading the file:%s", err)
		return nil, err
	}
	return result, nil
}
