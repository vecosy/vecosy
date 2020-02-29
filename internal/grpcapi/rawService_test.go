package grpcapi

import (
	"context"
	"fmt"
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

func TestServer_GetFile(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	for _, security := range []bool{false, true} {
		t.Run(fmt.Sprintf("GetFile_Security_%v", security), func(t *testing.T) {
			mockRepo := mocks.NewMockRepo(ctrl)
			srv, err := NewNoTLS(mockRepo, ":8080", security)
			check.NoError(err)
			check.NotNil(srv)
			app := configrepo.NewApplicationVersion("app", "1.0.0")

			filePath := "config.yml"
			repoFile := &configrepo.RepoFile{
				Version: uuid.New().String(),
				Content: []byte(uuid.New().String()),
			}
			mockRepo.EXPECT().GetFile(app, filePath).Return(repoFile, nil)
			request := &GetFileRequest{
				AppName:    app.AppName,
				AppVersion: app.AppVersion,
				FilePath:   filePath,
			}
			ctx := context.Background()
			if security {
				ctx = applySecurityIn(t, privKey, ctx, mockRepo, app.AppName, app.AppVersion)
			}
			response, err := srv.GetFile(ctx, request)
			check.NoError(err)
			check.NotNil(response)
			check.Equal(response.FileContent, repoFile.Content)
		})
	}
}

func TestServer_GetFile_NotFound(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	for _, security := range []bool{false, true} {
		t.Run(fmt.Sprintf("GetFile_NotFound_Security_%v", security), func(t *testing.T) {
			mockRepo := mocks.NewMockRepo(ctrl)
			srv, err := NewNoTLS(mockRepo, ":8080", security)
			check.NoError(err)
			check.NotNil(srv)
			app := configrepo.NewApplicationVersion("app", "1.0.0")
			filePath := "config.yml"

			notFoundError := fmt.Errorf("not found")
			mockRepo.EXPECT().GetFile(app, filePath).Return(nil, notFoundError)
			request := &GetFileRequest{
				AppName:    app.AppName,
				AppVersion: app.AppVersion,
				FilePath:   filePath,
			}
			ctx := context.Background()
			if security {
				ctx = applySecurityIn(t, privKey, ctx, mockRepo, app.AppName, app.AppVersion)
			}
			response, err := srv.GetFile(ctx, request)
			check.EqualError(err, notFoundError.Error())
			check.Nil(response)
		})
	}
}

func TestServer_GetFile_Unauthorized(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	privKey, _, err := utils.GenerateKeyPair()
	privKeyWrong, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	mockRepo := mocks.NewMockRepo(ctrl)
	srv, err := NewNoTLS(mockRepo, ":8080", true)
	check.NoError(err)
	check.NotNil(srv)
	app := configrepo.NewApplicationVersion("app", "1.0.0")
	filePath := "config.yml"

	request := &GetFileRequest{
		AppName:    app.AppName,
		AppVersion: app.AppVersion,
		FilePath:   filePath,
	}
	ctx := context.Background()
	md := metadata.MD{"token": []string{utils.GenJwsFromPrivateKey(t, privKeyWrong, "TestApp").FullSerialize()}}
	ctx = metadata.NewIncomingContext(ctx, md)
	prepareSecurityMock(app.AppName, app.AppVersion, mockRepo, privKey)

	response, err := srv.GetFile(ctx, request)
	check.Equal(err, security.AuthFailed)
	check.Nil(response)
}

func TestServer_GetFile_InvalidApp(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepo(ctrl)
	srv, err := NewNoTLS(mockRepo, ":8080", true)
	check.NoError(err)
	check.NotNil(srv)
	filePath := "config.yml"

	badAppNameRequest := &GetFileRequest{
		AppName:    "",
		AppVersion: "1.0.0",
		FilePath:   filePath,
	}
	response, err := srv.GetFile(context.Background(), badAppNameRequest)
	check.Equal(err, validation.InvalidApplicationName)
	check.Nil(response)

	badAppVersionRequest := &GetFileRequest{
		AppName:    "app1",
		AppVersion: "",
		FilePath:   filePath,
	}
	response, err = srv.GetFile(context.Background(), badAppVersionRequest)
	check.Equal(err, validation.InvalidVersion)
	check.Nil(response)
}
