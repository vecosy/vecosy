package rest

import (
	"fmt"
	"github.com/imdario/mergo"
	"github.com/kataras/iris"
	"github.com/kataras/iris/core/router"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"path"
	"regexp"
	"strings"
)

type propertySources struct {
	Name    string                 `json:"name"`
	Source  map[string]interface{} `json:"source"`
	version string
}

type springSummaryResponse struct {
	Name            string             `json:"name"`
	Profiles        []string           `json:"profiles"`
	Label           *string            `json:"label"`
	Version         string             `json:"version"`
	State           *string            `json:"state"`
	PropertySources []*propertySources `json:"propertySources"`
}

func (s *Server) registerSpringCloudEndpoints(parent router.Party) {
	springParty := parent.Party("/spring")
	springParty.Get("/{appVersion:string}/{appName:string}/{profile:string}", s.springAppInfo)
	springParty.Get("/{appVersion:string}/{appAndProfile:string}", s.springAppFile)
}

// GET:{appVersion:string}/{appName:string}/{profile:string}
func (s *Server) springAppInfo(ctx iris.Context) {
	appVersion := ctx.Params().GetString("appVersion")
	appName := ctx.Params().GetString("appName")
	profileParam := ctx.Params().GetString("profile")
	profiles := strings.Split(profileParam, ",")
	log := logrus.WithField("appName", appName).WithField("appVersion", appVersion).WithField("profiles", profiles)
	log.Info("springAppInfo")
	response := springSummaryResponse{
		Name:            appName,
		Profiles:        profiles,
		Label:           nil,
		State:           nil,
		PropertySources: make([]*propertySources, 0),
	}

	for _, configFilePath := range s.getApplicationFilePaths(appName, profiles) {
		propertySrc, err := s.getPropertySource(appName, appVersion, configFilePath)
		if err != nil {
			log.Errorf("Error getting resource:%s", err)
		} else {
			if propertySrc != nil {
				response.Version = propertySrc.version
				response.PropertySources = append(response.PropertySources, propertySrc)
			}
		}
	}

	_, err := ctx.JSON(response)
	if err != nil {
		log.Errorf("Error responding :%s", err)
		internalServerError(ctx)
	}
}

// GET: /{application}-{profile}.[yml|json]
func (s *Server) springAppFile(ctx iris.Context) {
	appVersion := ctx.Params().GetString("appVersion")
	appAndProfile := ctx.Params().GetString("appAndProfile")
	appName, ext, profile := extractAppNameAndVersion(appAndProfile)

	log := logrus.WithField("appName", appName).WithField("appVersion", appVersion)
	log = log.WithField("profiles", profile).WithField("extension", ext)
	log.Info("springAppFile")

	if ext != ".yml" && ext != ".json" && ext != ".yaml" {
		log.Errorf("Invalid extension :%s", ext)
		badRequest(ctx, "invalid extension, only json,yaml are supported")
		return
	}

	// reading and merging configurations
	finalConfig := make(map[interface{}]interface{})
	for _, configFilePath := range s.getApplicationFilePaths(appName, []string{profile}) {
		profileFile, err := s.repo.GetFile(appName, appVersion, configFilePath)
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

	// converting and responding
	var err error
	switch ext {
	case ".yml", ".yaml":
		_, err = ctx.YAML(finalConfig)
	case ".json":
		normalizedMap, err := utils.NormalizeMap(finalConfig)
		if err != nil {
			logrus.Errorf("Error normalizing json map:%#+vs, err:%s", finalConfig, err)
			internalServerError(ctx)
			return
		}
		_, err = ctx.JSON(normalizedMap)
	}
	if err != nil {
		log.Errorf("Error responding :%s", err)
		internalServerError(ctx)
	}
}

// Read a config file and convert it to propertySources
func (s *Server) getPropertySource(appName, appVersion, configFilePath string) (*propertySources, error) {
	profileFile, err := s.repo.GetFile(appName, appVersion, configFilePath)
	if err != nil {
		logrus.Warnf("Error getting file:%s, err:%s", configFilePath, err)
		return nil, err
	}

	// parsing the content
	config := make(map[interface{}]interface{})
	err = yaml.Unmarshal(profileFile.Content, config)
	if err != nil {
		logrus.Errorf("Error parsing yml file:%s, err:%s", configFilePath, err)
		return nil, err
	} else {
		resource := &propertySources{Name: configFilePath, Source: make(map[string]interface{})}
		resource.Source, err = utils.NormalizeMap(config)
		if err != nil {
			logrus.Errorf("Error normalizing json map:%#+vs, err:%s", config, err)
			return nil, err
		}
		resource.version = profileFile.Version
		return resource, nil
	}
}

var appProfileRe = regexp.MustCompile("([a-z|A-Z|0-9|.]*)*-?")

// Extract App Extension and Profile form a single string
// the string format is [app]-[profile].[extension]
// returns appName, extension, profile
func extractAppNameAndVersion(appAndProfile string) (string, string, string) {
	values := appProfileRe.FindAllStringSubmatch(appAndProfile, -1)
	logrus.Debugf("values %+v", values)
	if len(values) > 1 {
		appParts := make([]string, len(values)-1)
		for i := 0; i < len(values)-1; i++ {
			appParts[i] = values[i][1]
		}
		appName := strings.Join(appParts, "-")
		profileAndExtension := values[len(values)-1][1]
		ext := path.Ext(profileAndExtension)
		profile := strings.Replace(profileAndExtension, ext, "", 1)
		return appName, ext, profile
	} else {
		ext := path.Ext(appAndProfile)
		appName := strings.Replace(appAndProfile, ext, "", 1)
		return appName, ext, ""
	}
}

// Given an application name and a profile list, returns the related application file names
func (s *Server) getApplicationFilePaths(appName string, profiles []string) []string {
	appConfigFiles := make([]string, 0)
	appConfigFiles = append(appConfigFiles, getSpringCommonApplicationFile(appName))
	for _, profile := range profiles {
		if profile != "" {
			appConfigFiles = append(appConfigFiles, getSpringApplicationFile(appName, profile))
		}
	}
	return appConfigFiles
}

func getSpringApplicationFile(appName, profile string) string {
	return fmt.Sprintf("%s-%s.yml", appName, profile)
}

func getSpringCommonApplicationFile(appName string) string {
	return fmt.Sprintf("%s.yml", appName)
}
