package grpcapi

import (
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"
)

func StartGRPCServerIT(ctrl *gomock.Controller, t *testing.T, security bool) (*mocks.MockRepo, *Server) {
	mockRepo := mocks.NewMockRepo(ctrl)
	freePort, err := freeport.GetFreePort()
	assert.NoError(t, err)
	address := fmt.Sprintf("127.0.0.1:%d", freePort)
	srv, err := NewNoTLS(mockRepo, address, security)
	assert.NoError(t, err)
	go func() {
		err := srv.Start()
		if err != nil {
			assert.FailNow(t, "error starting grpc server %s", err)
		}
	}()
	time.Sleep(1 * time.Second)
	return mockRepo, srv
}

// used on the integration tests
func applySecurityOut(t *testing.T, privKey *rsa.PrivateKey, ctx context.Context, repo *mocks.MockRepo, appName, appVersion string) context.Context {
	prepareSecurityMock(appName, appVersion, repo, privKey)
	return metadata.AppendToOutgoingContext(ctx, "token", utils.GenJwsFromPrivateKey(t, privKey, "TestApp").FullSerialize())
}

// used on the unit tests
func applySecurityIn(t *testing.T, privKey *rsa.PrivateKey, ctx context.Context, repo *mocks.MockRepo, appName, appVersion string) context.Context {
	prepareSecurityMock(appName, appVersion, repo, privKey)
	md := metadata.MD{"token": []string{utils.GenJwsFromPrivateKey(t, privKey, "TestApp").FullSerialize()}}
	return metadata.NewIncomingContext(ctx, md)
}

func prepareSecurityMock(appName string, appVersion string, repo *mocks.MockRepo, privKey *rsa.PrivateKey) {
	cfgApp := &configrepo.ApplicationVersion{AppName: appName, AppVersion: appVersion}
	repo.EXPECT().GetFile(cfgApp, "pub.key").Return(&configrepo.RepoFile{
		Version: uuid.New().String(),
		Content: utils.PublicKeyToBytes(&privKey.PublicKey),
	}, nil)
}
