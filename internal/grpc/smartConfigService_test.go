package grpc

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"testing"
)

func TestServer_GetConfig(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepo(ctrl)
	srv := New(mockRepo, ":8080")
	check.NotNil(srv)
	appName := "app"
	appVersion := "1.0.0"
	env := "dev"
	devContent := `environment: dev`
	commonContent := `version: 1.0.0`
	repoVersion := uuid.New().String()
	devRepoFile := &configrepo.RepoFile{Version: repoVersion, Content: []byte(devContent)}
	commonRepoFile := &configrepo.RepoFile{Version: repoVersion, Content: []byte(commonContent)}
	mockRepo.EXPECT().GetFile(appName, appVersion, "dev/config.yml").Return(devRepoFile, nil)
	mockRepo.EXPECT().GetFile(appName, appVersion, "config.yml").Return(commonRepoFile, nil)
	request := &GetConfigRequest{
		AppName:     appName,
		AppVersion:  appVersion,
		Environment: env,
	}
	response, err := srv.GetConfig(context.Background(), request)
	check.NoError(err)
	check.NotNil(response)
	appConfig := make(map[string]string)
	check.NoError(yaml.Unmarshal([]byte(response.ConfigContent), &appConfig))
	expectedConfig := map[string]string{
		"environment": "dev",
		"version":     "1.0.0",
	}
	check.Equal(expectedConfig, appConfig)
}

func TestServer_GetConfig_NotFound(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepo(ctrl)
	srv := New(mockRepo, ":8080")
	check.NotNil(srv)
	appName := "app"
	appVersion := "1.0.0"
	env := "dev"
	mockRepo.EXPECT().GetFile(appName, appVersion, "config.yml").Return(nil, configrepo.ApplicationNotFoundError)
	request := &GetConfigRequest{
		AppName:     appName,
		AppVersion:  appVersion,
		Environment: env,
	}
	response, err := srv.GetConfig(context.Background(), request)
	check.EqualError(err, configrepo.ApplicationNotFoundError.Error())
	check.Nil(response)
}
