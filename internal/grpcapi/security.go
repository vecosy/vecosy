package grpcapi

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/security"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"google.golang.org/grpc/metadata"
)

// CheckToken checks if the request has a valid token on the GRPC metadata
func (s *Server) CheckToken(ctx context.Context, app *configrepo.ApplicationVersion) error {
	log := logrus.WithField("method", "CheckToken")
	if !s.IsSecurityEnabled() {
		return nil
	}
	md, found := metadata.FromIncomingContext(ctx)
	if !found {
		return security.ErrNoMetadataFound
	}
	tokens := md.Get("token")
	if len(tokens) != 1 {
		return security.ErrAuthFailed
	}
	token := tokens[0]
	log.Debugf("metadata token:%s", token)
	err := security.CheckJwtToken(s.repo, app, token)
	if err != nil {
		return err
	}
	return nil
}
