package grpcapi

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/merger"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/internal/validation"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

var smartConfigFileMerger = merger.SmartConfigMerger{}

func (s *Server) GetConfig(ctx context.Context, request *GetConfigRequest) (*GetConfigResponse, error) {
	log := logrus.WithField("method", "GRPC:GetConfig").WithField("request", request)
	log.Infof("GetConfig")

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

	config, err := smartConfigFileMerger.Merge(s.repo, appVersion, []string{request.Environment})
	if err != nil {
		log.Errorf("error merging smartconfig:%s", err)
		return nil, err
	}

	normConfig, err := utils.NormalizeMap(config)
	if err != nil {
		log.Errorf("error normalizing config:%s", err)
		return nil, err
	}

	yml, err := yaml.Marshal(normConfig)
	logrus.Debugf("received:%s", string(yml))
	if err != nil {
		log.Errorf("error generating yaml:%s", err)
		return nil, err
	}
	return &GetConfigResponse{
		ConfigContent: string(yml),
	}, nil
}
