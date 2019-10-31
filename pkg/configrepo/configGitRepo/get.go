package configGitRepo

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io/ioutil"
)

func (cr *GitConfigRepo) GetNearestBranch(targetApp, targetVersion string) (*plumbing.Reference, error) {
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

func (cr *GitConfigRepo) GetFile(targetApp, targetVersion, path string) (*configrepo.RepoFile, error) {
	log := logrus.WithField("method", "GetFile").WithField("targetApp", targetApp).WithField("targetVersion", targetVersion).WithField("path", path)
	branchRef, err := cr.GetNearestBranch(targetApp, targetVersion)
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
