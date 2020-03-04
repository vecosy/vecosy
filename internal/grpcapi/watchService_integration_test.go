// +build integration

package grpcapi

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/testutil"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"google.golang.org/grpc"
	"io"
	"testing"
	"time"
)

func TestServer_Watch_IT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	privKey, _, err := testutil.GenerateKeyPair()
	assert.NoError(t, err)

	mockRepo, srv := startGRPCServerIT(ctrl, t, true)

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

	onChangeCh := make(chan configrepo.OnChangeHandler, 1)
	mockRepo.EXPECT().AddOnChangeHandler(gomock.Any()).Do(func(handler configrepo.OnChangeHandler) {
		onChangeCh <- handler
	})

	ctx := context.Background()
	ctx = applySecurityOut(ctx, t, privKey, mockRepo, app.AppName, app.AppVersion)
	stream, err := cl.Watch(ctx, request)
	assert.NoError(t, err)

	var onChangeHandlerCapture configrepo.OnChangeHandler
	timeout := time.NewTimer(1 * time.Second)
	select {
	case onChangeHandlerCapture = <-onChangeCh:
		t.Logf("onChange handler captured")
	case <-timeout.C:
		assert.FailNow(t, "timeout occurred")

	}

	// waiting for a message
	recvWatchCh := make(chan *WatchResponse)
	go func() {
		for {
			response, err := stream.Recv()
			if err == io.EOF {
				return
			}
			assert.NotNil(t, response)
			logrus.Debugf("Received %+v", *response)
			assert.NoError(t, err)
			recvWatchCh <- response
			return
		}
	}()

	//simulate repo changes
	assert.NotNil(t, onChangeHandlerCapture)
	onChangeHandlerCapture(configrepo.ApplicationVersion{AppName: app.AppName, AppVersion: app.AppVersion})

	timeout.Reset(1 * time.Second)
	select {
	case <-timeout.C:
		assert.FailNow(t, "timeout occurred")
	case changed := <-recvWatchCh:
		logrus.Debugf("Received changes :%+v", *changed)
	}
	srv.Stop()
}
