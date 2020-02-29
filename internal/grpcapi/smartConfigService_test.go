package grpcapi

import (
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/security"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/internal/validation"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestServer_GetConfig(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	for _, security := range []bool{false, true} {
		t.Run(fmt.Sprintf("GetConfig_Security_%v", security), func(t *testing.T) {

			mockRepo := mocks.NewMockRepo(ctrl)
			srv, err := NewNoTLS(mockRepo, ":8080", security)
			check.NoError(err)
			check.NotNil(srv)
			app := configrepo.NewApplicationVersion("app", "1.0.0")
			env := "dev"
			devContent := `environment: dev`
			commonContent := `version: 1.0.0`
			repoVersion := uuid.New().String()
			devRepoFile := &configrepo.RepoFile{Version: repoVersion, Content: []byte(devContent)}
			commonRepoFile := &configrepo.RepoFile{Version: repoVersion, Content: []byte(commonContent)}
			mockRepo.EXPECT().GetFile(app, "dev/config.yml").Return(devRepoFile, nil)
			mockRepo.EXPECT().GetFile(app, "config.yml").Return(commonRepoFile, nil)
			request := &GetConfigRequest{
				AppName:     app.AppName,
				AppVersion:  app.AppVersion,
				Environment: env,
			}
			ctx := context.Background()
			if security {
				ctx = applySecurityIn(t, privKey, ctx, mockRepo, app.AppName, app.AppVersion)
			}
			response, err := srv.GetConfig(ctx, request)
			check.NoError(err)
			check.NotNil(response)
			appConfig := make(map[string]string)
			check.NoError(yaml.Unmarshal([]byte(response.ConfigContent), &appConfig))
			expectedConfig := map[string]string{
				"environment": "dev",
				"version":     "1.0.0",
			}
			check.Equal(expectedConfig, appConfig)
		})
	}
}

func TestServer_GetConfig_NotFound(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	for _, security := range []bool{false, true} {
		t.Run(fmt.Sprintf("GetConfig_NotFound_Security_%v", security), func(t *testing.T) {
			mockRepo := mocks.NewMockRepo(ctrl)
			srv, err := NewNoTLS(mockRepo, ":8080", security)
			check.NoError(err)
			check.NotNil(srv)
			app := configrepo.NewApplicationVersion("app", "1.0.0")
			env := "dev"

			mockRepo.EXPECT().GetFile(app, "config.yml").Return(nil, configrepo.ApplicationNotFoundError)
			request := &GetConfigRequest{
				AppName:     app.AppName,
				AppVersion:  app.AppVersion,
				Environment: env,
			}
			ctx := context.Background()
			if security {
				ctx = applySecurityIn(t, privKey, ctx, mockRepo, app.AppName, app.AppVersion)
			}
			response, err := srv.GetConfig(ctx, request)
			check.EqualError(err, configrepo.ApplicationNotFoundError.Error())
			check.Nil(response)
		})
	}
}

func TestServer_GetConfig_Unauthorized(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)
	privKeyWrong, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	mockRepo := mocks.NewMockRepo(ctrl)
	srv, err := NewNoTLS(mockRepo, ":8080", true)
	check.NoError(err)
	check.NotNil(srv)
	app := configrepo.NewApplicationVersion("app", "1.0.0")
	env := "dev"

	request := &GetConfigRequest{
		AppName:     app.AppName,
		AppVersion:  app.AppVersion,
		Environment: env,
	}
	ctx := context.Background()
	md := metadata.MD{"token": []string{utils.GenJwsFromPrivateKey(t, privKeyWrong, "TestApp").FullSerialize()}}
	ctx = metadata.NewIncomingContext(ctx, md)
	prepareSecurityMock(app.AppName, app.AppVersion, mockRepo, privKey)

	response, err := srv.GetConfig(ctx, request)
	check.Equal(err, security.AuthFailed)
	check.Nil(response)
}

func TestServer_GetConfig_InvalidApp(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepo(ctrl)
	srv, err := NewNoTLS(mockRepo, ":8080", true)
	check.NoError(err)
	check.NotNil(srv)

	badAppNameRequest := &GetConfigRequest{
		AppName:     "",
		AppVersion:  "1.0.0",
		Environment: "dev",
	}
	response, err := srv.GetConfig(context.Background(), badAppNameRequest)
	check.Equal(err, validation.InvalidApplicationName)
	check.Nil(response)

	badAppVersionRequest := &GetConfigRequest{
		AppName:     "app1",
		AppVersion:  "",
		Environment: "dev",
	}
	response, err = srv.GetConfig(context.Background(), badAppVersionRequest)
	check.Equal(err, validation.InvalidVersion)
	check.Nil(response)
}
