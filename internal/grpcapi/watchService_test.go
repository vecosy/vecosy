package grpcapi

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/security"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/internal/validation"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"
)

func TestServer_Watch(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	for _, security := range []bool{false, true} {
		t.Run(fmt.Sprintf("Watch_Security_%v", security), func(t *testing.T) {
			mockRepo := mocks.NewMockRepo(ctrl)
			srv, err := NewNoTLS(mockRepo, ":8080", security)
			check.NoError(err)
			check.NotNil(srv)
			appName := "app"
			appVersion := "1.0.0"

			onChangeCh := make(chan configrepo.OnChangeHandler, 1)
			mockRepo.EXPECT().AddOnChangeHandler(gomock.Any()).Do(func(handler configrepo.OnChangeHandler) {
				onChangeCh <- handler
			})

			app := &Application{
				AppName:    appName,
				AppVersion: appVersion,
			}
			request := &WatchRequest{
				WatcherName: "test",
				Application: app,
			}
			stream := NewMockWatchService_WatchServer(ctrl)
			streamCtx, _ := context.WithTimeout(context.Background(), 1*time.Second)
			if security {
				streamCtx = applySecurityIn(t, privKey, streamCtx, mockRepo, app.AppName, app.AppVersion)
			}
			stream.EXPECT().Context().AnyTimes().Return(streamCtx)
			err = srv.Watch(request, stream)
			check.NoError(err)
			check.NotEmpty(onChangeCh)
			capturedHandler := <-onChangeCh
			check.NotNil(capturedHandler)
		})
	}
}

func TestServer_Watch_Unauthorized(t *testing.T) {
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
	appName := "app"
	appVersion := "1.0.0"

	app := &Application{
		AppName:    appName,
		AppVersion: appVersion,
	}
	request := &WatchRequest{
		WatcherName: "test",
		Application: app,
	}
	stream := NewMockWatchService_WatchServer(ctrl)
	streamCtx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	md := metadata.MD{"token": []string{utils.GenJwsFromPrivateKey(t, privKeyWrong, "TestApp").FullSerialize()}}
	streamCtx = metadata.NewIncomingContext(streamCtx, md)
	prepareSecurityMock(app.AppName, app.AppVersion, mockRepo, privKey)
	stream.EXPECT().Context().AnyTimes().Return(streamCtx)
	err = srv.Watch(request, stream)
	check.Equal(err, security.AuthFailed)
}

func TestServer_Watch_InvalidApplication(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepo(ctrl)
	srv, err := NewNoTLS(mockRepo, ":8080", true)
	check.NoError(err)
	check.NotNil(srv)

	badAppNameRequest := &WatchRequest{
		WatcherName: "test",
		Application: &Application{
			AppName:    "",
			AppVersion: "1.0.0",
		},
	}
	stream := NewMockWatchService_WatchServer(ctrl)
	err = srv.Watch(badAppNameRequest, stream)
	check.Equal(err, validation.InvalidApplicationName)

	badAppVersionRequest := &WatchRequest{
		WatcherName: "test",
		Application: &Application{
			AppName:    "app1",
			AppVersion: "",
		},
	}
	stream = NewMockWatchService_WatchServer(ctrl)
	err = srv.Watch(badAppVersionRequest, stream)
	check.Equal(err, validation.InvalidVersion)
}
