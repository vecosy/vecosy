// +build integration

package grpcapi

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"google.golang.org/grpc"
	"testing"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	m.Run()
}

func TestServer_GetFile_IT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo, srv := StartGRPCServerIT(ctrl, t, false)

	appName := "app1"
	appVersion := "1.0.0"
	app := configrepo.NewApplicationVersion(appName, appVersion)

	fileName := "config.yaml"
	fileContent := []byte(uuid.New().String())

	mockRepo.EXPECT().GetFile(app, fileName).Return(&configrepo.RepoFile{
		Version: uuid.New().String(),
		Content: fileContent,
	}, nil)
	conn, err := grpc.Dial(srv.address, grpc.WithInsecure())
	assert.NoError(t, err)
	defer conn.Close()
	cl := NewRawClient(conn)
	req := &GetFileRequest{
		AppName:    appName,
		AppVersion: appVersion,
		FilePath:   fileName,
	}
	resp, err := cl.GetFile(context.TODO(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.FileContent, fileContent)
	srv.Stop()
}
