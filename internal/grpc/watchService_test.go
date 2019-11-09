package grpc

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"testing"
	"time"
)

func TestServer_Watch(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepo(ctrl)
	srv := New(mockRepo, ":8080")
	check.NotNil(srv)
	appName := "app"
	appVersion := "1.0.0"

	onChangeCh := make(chan configrepo.OnChangeHandler, 1)
	mockRepo.EXPECT().AddOnChangeHandler(gomock.Any()).Do(func(handler configrepo.OnChangeHandler) {
		onChangeCh <- handler
	})

	request := &WatchRequest{
		WatcherName: "test",
		Application: &Application{
			AppName:    appName,
			AppVersion: appVersion,
		},
	}
	stream := NewMockWatchService_WatchServer(ctrl)
	streamCtx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	stream.EXPECT().Context().Return(streamCtx)
	err := srv.Watch(request, stream)
	check.NoError(err)
	check.NotEmpty(onChangeCh)
	capturedHandler := <-onChangeCh
	check.NotNil(capturedHandler)
}
