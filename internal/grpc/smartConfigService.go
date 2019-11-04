package grpc

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/merger"
)

var smartConfigFileMerger = merger.SmartConfigMerger{}

func (s *Server) GetConfig(ctx context.Context, request *GetConfigRequest) (*GetConfigResponse, error) {
	log := logrus.WithField("method", "GRPC:GetConfig").WithField("request", request)
	log.Infof("GetConfig")
	config, err := smartConfigFileMerger.Merge(s.repo, request.AppName, request.AppVersion, []string{request.Environment})
	if err != nil {
		log.Errorf("error merging smartconfig:%s", err)
		return nil, err
	}
	yml, err := yaml.Marshal(config)
	if err != nil {
		log.Errorf("error generating yaml:%s", err)
		return nil, err
	}
	return &GetConfigResponse{
		ConfigContent: string(yml),
	}, nil
}
