package configrepo

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/vconf/v2/internal/utils"
	"gopkg.in/src-d/go-git.v4"
	plumbing2 "gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"sort"
)

type ConfigRepo struct {
	repo     *git.Repository
	branches map[string]*plumbing2.Reference
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
	return &ConfigRepo{repo, make(map[string]*plumbing2.Reference)}, nil
}

func (cr *ConfigRepo) Init() error {
	return cr.LoadBranches()
}

func (cr *ConfigRepo) LoadBranches() error {
	result := make(map[string]*plumbing2.Reference)
	branches, err := cr.repo.Branches()
	if err != nil {
		return nil
	}
	err = branches.ForEach(func(branchRef *plumbing2.Reference) error {
		result[branchRef.Name().String()] = branchRef
		return nil
	})
	if err != nil {
		return nil
	}
	cr.branches = result
	return nil
}

func (cr *ConfigRepo) GetNearestBranch(targetVersion string) (*plumbing2.Reference, error) {
	branchVersions := make([]*version.Version, len(cr.branches), len(cr.branches))
	i := 0
	for branchName, _ := range cr.branches {
		branchVer, verErr := version.NewVersion(branchName)
		if verErr != nil {
			logrus.Warn("Branch name is not a valid version %s err:%s", branchName, verErr)
		}
		branchVersions[i] = branchVer
		i++
	}
	sort.Sort(version.Collection(branchVersions))
	utils.ReverseVersion(branchVersions)

	constraint, err := version.NewConstraint(fmt.Sprintf("<=%s", targetVersion))
	if err != nil {
		return nil, err
	}
	for _, chkVer := range branchVersions {
		if constraint.Check(chkVer) {
			return cr.branches[chkVer.String()], nil
		}
	}
	return nil, fmt.Errorf("no branch found for target chkVer:%s", targetVersion)
}

func (cr *ConfigRepo) GetFile(targetVersion, path string) (*object.File, error) {
	branchRef, err := cr.GetNearestBranch(targetVersion)
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
