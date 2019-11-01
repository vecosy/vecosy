package merger

import (
	"fmt"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

type SpringMerger struct{}

func (m SpringMerger) Merge(repo configrepo.Repo, appName, appVersion string, profiles []string) (map[interface{}]interface{}, error) {
	// reading and merging configurations
	appConfigFiles := GetSpringApplicationFilePaths(appName, profiles, true)
	finalConfig := mergeFiles(repo, appName, appVersion, appConfigFiles)
	return finalConfig, nil
}

func GetSpringApplicationFilePaths(appName string, profiles []string, commonFirst bool) []string {
	appConfigFiles := make([]string, 0)
	appConfigFiles = append(appConfigFiles, getSpringCommonApplicationFile("application"))
	appConfigFiles = append(appConfigFiles, getSpringCommonApplicationFile(appName))
	for _, profile := range profiles {
		if profile != "" {
			appConfigFiles = append(appConfigFiles, getSpringApplicationFile("application", profile))
			appConfigFiles = append(appConfigFiles, getSpringApplicationFile(appName, profile))
		}
	}
	if !commonFirst {
		utils.ReverseStrings(appConfigFiles)
	}
	return appConfigFiles
}

func getSpringApplicationFile(appName, profile string) string {
	return fmt.Sprintf("%s-%s.yml", appName, profile)
}

func getSpringCommonApplicationFile(appName string) string {
	return fmt.Sprintf("%s.yml", appName)
}
