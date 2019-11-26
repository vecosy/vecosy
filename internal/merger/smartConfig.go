package merger

import (
	"fmt"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

type SmartConfigMerger struct{}

func (s SmartConfigMerger) Merge(repo configrepo.Repo, app *configrepo.ApplicationVersion, profiles []string) (map[interface{}]interface{}, error) {
	appConfigFiles := getSmartConfigApplicationFilePaths(app.AppName, profiles)
	return mergeFiles(repo, app, appConfigFiles)
}

func getSmartConfigApplicationFilePaths(appName string, profiles []string) []string {
	appConfigFiles := make([]string, 1)
	appConfigFiles[0] = getSmartConfigCommonApplicationFile()
	for _, profile := range profiles {
		if profile != "" {
			appConfigFiles = append(appConfigFiles, getSmartConfigApplicationFile(profile))
		}
	}
	return appConfigFiles
}

func getSmartConfigApplicationFile(profile string) string {
	return fmt.Sprintf("%s/config.yml", profile)
}

func getSmartConfigCommonApplicationFile() string {
	return "config.yml"
}
