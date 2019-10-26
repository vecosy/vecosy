package configGitRepo

import (
	"fmt"
	"github.com/hashicorp/go-version"
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

func (cr *GitConfigRepo) GetFile(targetApp, targetVersion, path string) ([]byte, error) {
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
	fl, err := tree.File(path)
	if err != nil {
		return nil, err
	}
	flReader, err := fl.Reader()
	if err != nil {
		return nil, err
	}
	defer flReader.Close()
	return ioutil.ReadAll(flReader)
}
