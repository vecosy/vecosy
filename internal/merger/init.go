package merger

import (
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"gopkg.in/yaml.v2"
)

type ConfigMerger interface {
	Merge(repo configrepo.Repo, appName, appVersion string, profiles []string) (map[interface{}]interface{}, error)
}

func mergeFiles(repo configrepo.Repo, app *configrepo.ApplicationVersion, appConfigFiles []string) (map[interface{}]interface{}, error) {
	finalConfig := make(map[interface{}]interface{})
	for _, configFilePath := range appConfigFiles {
		profileFile, err := repo.GetFile(app, configFilePath)
		if err != nil {
			if err == configrepo.ApplicationNotFoundError {
				return nil, err
			}
			logrus.Warnf("Error getting file:%s, err:%s", configFilePath, err)
		} else {
			fileConfig := make(map[interface{}]interface{})
			err = yaml.Unmarshal(profileFile.Content, fileConfig)
			if err != nil {
				return nil, errors.Wrapf(err, "Error parsing yml file:%s, err:%s", configFilePath, err)
			}
			err = mergo.Map(&finalConfig, fileConfig, mergo.WithOverride)
			if err != nil {
				return nil, errors.Wrapf(err, "Error merging configuration :%#+v, err:%s", fileConfig, err)
			}
		}
	}
	return finalConfig, nil
}
