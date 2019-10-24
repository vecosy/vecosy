package configrepo

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func (cr *ConfigRepo) GetNearestBranch(targetApp, targetVersion string) (*plumbing.Reference, error) {
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
