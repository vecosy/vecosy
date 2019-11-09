package grpc

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/mocks"
	"testing"
	"time"
)

func StartGRPCServerIT(ctrl *gomock.Controller, t *testing.T) (*mocks.MockRepo, *Server) {
	mockRepo := mocks.NewMockRepo(ctrl)
	freePort, err := freeport.GetFreePort()
	assert.NoError(t, err)
	address := fmt.Sprintf("127.0.0.1:%d", freePort)
	srv := New(mockRepo, address)
	go func() {
		err := srv.Start()
		if err != nil {
			assert.FailNow(t, "error starting grpc server %s", err)
		}
	}()
	time.Sleep(1 * time.Second)
	return mockRepo, srv
}
