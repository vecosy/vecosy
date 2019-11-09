package grpc

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"testing"
)

func TestServer_GetFile(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepo(ctrl)
	srv := New(mockRepo, ":8080")
	check.NotNil(srv)
	appName := "app"
	appVersion := "1.0.0"
	filePath := "config.yml"
	repoFile := &configrepo.RepoFile{
		Version: uuid.New().String(),
		Content: []byte(uuid.New().String()),
	}
	mockRepo.EXPECT().GetFile(appName, appVersion, filePath).Return(repoFile, nil)
	request := &GetFileRequest{
		AppName:    appName,
		AppVersion: appVersion,
		FilePath:   filePath,
	}
	response, err := srv.GetFile(context.TODO(), request)
	check.NoError(err)
	check.NotNil(response)
	check.Equal(response.FileContent, repoFile.Content)
}

func TestServer_GetFile_NotFound(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepo(ctrl)
	srv := New(mockRepo, ":8080")
	check.NotNil(srv)
	appName := "app"
	appVersion := "1.0.0"
	filePath := "config.yml"

	notFoundError := fmt.Errorf("not found")
	mockRepo.EXPECT().GetFile(appName, appVersion, filePath).Return(nil, notFoundError)
	request := &GetFileRequest{
		AppName:    appName,
		AppVersion: appVersion,
		FilePath:   filePath,
	}
	response, err := srv.GetFile(context.TODO(), request)
	check.EqualError(err, notFoundError.Error())
	check.Nil(response)
}
