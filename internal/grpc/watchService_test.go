package vconf

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestServer_Watch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo, srv := initGrpc(ctrl, t)

	conn, err := grpc.Dial(srv.address, grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()
	cl := NewWatchServiceClient(conn)

	app := &Application{
		AppName:    "app1",
		AppVersion: "1.0.0",
	}
	request := &WatchRequest{
		WatcherName: "testWatcher",
		Application: app,
	}

	var onChangeHandlerCapture configrepo.OnChangeHandler
	mockRepo.EXPECT().AddOnChangeHandler(gomock.Any()).Do(func(handler configrepo.OnChangeHandler) {
		onChangeHandlerCapture = handler
	})

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	stream, err := cl.Watch(ctx, request)
	assert.NoError(t, err)
	time.Sleep(2 * time.Second)

	recvWatchCh := make(chan *WatchResponse)
	go func() {
		for {
			response, err := stream.Recv()
			assert.NoError(t, err)
			recvWatchCh <- response
		}
	}()
	assert.NotNil(t, onChangeHandlerCapture)

	onChangeHandlerCapture(app.AppName, app.AppVersion)

	timeout := time.NewTimer(1 * time.Second).C
	select {
	case <-timeout:
		assert.FailNow(t, "timeout occured")
	case changed := <-recvWatchCh:
		t.Logf("Received changes :%+v", changed)
	}
	srv.Stop()
}
