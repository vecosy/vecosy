package grpcapi

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/validation"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

func (s *Server) GetFile(ctx context.Context, request *GetFileRequest) (*GetFileResponse, error) {
	log := logrus.WithField("method", "GetFile").WithField("request", request)
	appVersion := configrepo.NewApplicationVersion(request.AppName, request.AppVersion)
	err := validation.ValidateApplicationVersion(appVersion)
	if err != nil {
		log.Errorf("Error validating the application:%+v", appVersion)
		return nil, err
	}
	err = s.CheckToken(ctx, appVersion)
	if err != nil {
		log.Errorf("Error checking token:%s", err)
		return nil, err
	}

	file, err := s.repo.GetFile(appVersion, request.FilePath)
	if err != nil {
		return nil, err
	}
	return &GetFileResponse{
		FileContent: file.Content,
	}, nil
}
