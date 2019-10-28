package rest

import (
	"fmt"
	yamlJson "github.com/ghodss/yaml"
	"github.com/kataras/iris"
	"github.com/kataras/iris/core/router"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"path"
	"regexp"
	"strings"
)

type propertySources struct {
	Name   string                 `json:"name"`
	Source map[string]interface{} `json:"source"`
}

type springSummaryResponse struct {
	Name            string            `json:"name"`
	Profiles        []string          `json:"profiles"`
	Label           *string           `json:"label"`
	Version         string            `json:"version"`
	State           *string           `json:"state"`
	PropertySources []propertySources `json:"propertySources"`
}

/**
/{application}/{profile}[/{label}]
/{application}-{profile}.yml
/{label}/{application}-{profile}.yml

/{application}-{profile}.properties
/{label}/{application}-{profile}.properties
*/
func (s *Server) registerSpringCloudEndpoints(parent router.Party) {
	parent.Get("/{appVersion:string}/{appName:string}/{profile:string}", s.getSpringFileByAppAndProfile)
	parent.Get("/{appVersion:string}/{appAndProfile:string}", s.getSpringYamlByAppAndProfile)
}

// //{appVersion:string}/{appName:string}/{profile:string}/{application}/{profile}
func (s *Server) getSpringFileByAppAndProfile(ctx iris.Context) {
	appVersion := ctx.Params().GetString("appVersion")
	appName := ctx.Params().GetString("appName")
	profileParam := ctx.Params().GetString("profile")
	profiles := strings.Split(profileParam, ",")
	log := logrus.WithField("appName", appName).WithField("appVersion", appVersion).WithField("profiles", profiles)
	log.Info("getSpringFileByAppAndProfile")
	response := springSummaryResponse{
		Name:            appName,
		Profiles:        profiles,
		Label:           nil,
		State:           nil,
		PropertySources: make([]propertySources, 0),
	}

	for _, configFilePath := range s.getApplicationFilePaths(appName, profiles) {
		if configFilePath != "" {
			profileFile, err := s.repo.GetFile(appName, appVersion, configFilePath)
			if err != nil {
				logrus.Warnf("Error getting file:%s, err:%s", configFilePath, err)
			} else {
				resource := propertySources{Name: configFilePath, Source: make(map[string]interface{})}
				err = yaml.Unmarshal(profileFile.Content, resource.Source)
				if err != nil {
					logrus.Errorf("Error parsing yml file:%s, err:%s", configFilePath, err)
				} else {
					response.Version = profileFile.Version
					response.PropertySources = append(response.PropertySources, resource)
				}
			}
		}
	}
	_, err := ctx.JSON(response)
	if err != nil {
		log.Errorf("Error responding :%s", err)
		internalServerError(ctx)
	}
}

var appProfileRe = regexp.MustCompile("([a-z|A-Z|0-9|.]*)*-?")

func (s *Server) getSpringYamlByAppAndProfile(ctx iris.Context) {
	appVersion := ctx.Params().GetString("appVersion")
	appAndProfile := ctx.Params().GetString("appAndProfile")
	appName, ext, profile := extractAppNameAndVersion(appAndProfile)

	log := logrus.WithField("appName", appName).WithField("appVersion", appVersion).WithField("profiles", profile)
	log.Info("getSpringYamlByAppAndProfile")
	finalConfig := make(map[string]interface{})
	for _, configFilePath := range s.getApplicationFilePaths(appName, []string{profile}) {
		if configFilePath != "" {
			profileFile, err := s.repo.GetFile(appName, appVersion, configFilePath)
			if err != nil {
				logrus.Warnf("Error getting file:%s, err:%s", configFilePath, err)
			} else {
				err = yaml.Unmarshal(profileFile.Content, finalConfig)
				if err != nil {
					logrus.Errorf("Error parsing yml file:%s, err:%s", configFilePath, err)
				}
			}
		}
	}
	yml, err := yaml.Marshal(finalConfig)
	if err != nil {
		log.Errorf("Error creating yaml:%s", err)
		internalServerError(ctx)
		return
	}

	var response string
	switch ext {
	case ".yml":
		response = string(yml)
		break
	case ".json":
		jsonVal, err := yamlJson.YAMLToJSON(yml)
		if err != nil {
			log.Errorf("Error creating json:%s", err)
			internalServerError(ctx)
			return
		}
		response = string(jsonVal)
		break
	default:
		badRequest(ctx, "invalid extension, only json,yaml are supported")
	}
	_, err = ctx.WriteString(response)

	if err != nil {
		log.Errorf("Error responding :%s", err)
		internalServerError(ctx)
	}
}

func extractAppNameAndVersion(appAndProfile string) (string, string, string) {
	values := appProfileRe.FindAllStringSubmatch(appAndProfile, -1)
	logrus.Debugf("values %+v", values)
	appParts := make([]string, len(values)-1)
	for i := 0; i < len(values)-1; i++ {
		appParts[i] = values[i][1]
	}
	appName := strings.Join(appParts, "-")
	profileAndExtension := values[len(values)-1][1]
	ext := path.Ext(profileAndExtension)
	profile := strings.Replace(profileAndExtension, ext, "", 1)
	return appName, ext, profile
}

func (s *Server) getApplicationFilePaths(appName string, profiles []string) []string {
	appConfigFiles := make([]string, 0)
	for _, profile := range profiles {
		appConfigFiles = append(appConfigFiles, getSpringApplicationFile(appName, profile))
	}
	return append(appConfigFiles, getSpringCommonApplicationFile(appName))
}

func getSpringApplicationFile(appName, profile string) string {
	return fmt.Sprintf("%s-%s.yml", appName, profile)
}

func getSpringCommonApplicationFile(appName string) string {
	return fmt.Sprintf("%s.yml", appName)
}
