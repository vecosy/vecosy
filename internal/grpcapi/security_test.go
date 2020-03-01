package grpcapi

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/security"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestServer_CheckToken_NoMetadata(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepo(ctrl)
	srv, err := NewNoTLS(mockRepo, ":8080", true)
	check.NoError(err)
	ctxWithoutMeta := context.TODO()
	err = srv.CheckToken(ctxWithoutMeta, configrepo.NewApplicationVersion("app", "v1.0.0"))
	check.True(errors.Is(err, security.ErrNoMetadataFound))
}

func TestServer_CheckToken_NoTokenInMetadata(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepo(ctrl)
	srv, err := NewNoTLS(mockRepo, ":8080", true)
	check.NoError(err)
	ctx := context.Background()
	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("notValidTokenHeader", "notValidValue"))
	err = srv.CheckToken(ctx, configrepo.NewApplicationVersion("app", "v1.0.0"))
	check.True(errors.Is(err, security.ErrAuthFailed))
}
