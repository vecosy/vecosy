package merger

import (
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"gopkg.in/yaml.v2"
)

type ConfigMerger interface {
	Merge(repo configrepo.Repo, appName, appVersion string, profiles []string) (map[interface{}]interface{}, error)
}

func mergeFiles(repo configrepo.Repo, appName string, appVersion string, appConfigFiles []string, ) map[interface{}]interface{} {
	finalConfig := make(map[interface{}]interface{})
	for _, configFilePath := range appConfigFiles {
		profileFile, err := repo.GetFile(appName, appVersion, configFilePath)
		if err != nil {
			logrus.Warnf("Error getting file:%s, err:%s", configFilePath, err)
		} else {
			fileConfig := make(map[interface{}]interface{})
			err = yaml.Unmarshal(profileFile.Content, fileConfig)
			if err != nil {
				logrus.Errorf("Error parsing yml file:%s, err:%s", configFilePath, err)
			}
			err = mergo.Map(&finalConfig, fileConfig, mergo.WithOverride)
			if err != nil {
				logrus.Errorf("Error merging configuration :%#+v, err:%s", fileConfig, err)
			}
		}
	}
	return finalConfig
}
