package rest

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/merger"
)

var smartConfigFileMerger = merger.SmartConfigMerger{}

func (s *Server) registerSmartConfigEndpoints(parent iris.Party) {
	configApi := parent.Party("/config")
	configApi.Get("/", s.Info)
	configApi.Get("/{appName:string}/{appVersion:string}/{profile:string}", s.GetSmartConfig)
}

func (s *Server) GetSmartConfig(ctx iris.Context) {
	appName := ctx.Params().GetString("appName")
	appVersion := ctx.Params().GetString("appVersion")
	profile := ctx.Params().GetString("profile")
	requestedTypes := getAccepts(ctx)
	log := logrus.WithField("appName", appName).WithField("appVersion", appVersion).WithField("profile", profile)
	log = log.WithField("requested types", requestedTypes)
	log.Info("GetSmartConfig")
	var ext string
	if requestedTypes[context.ContentJSONHeaderValue] || requestedTypes[context.ContentHTMLHeaderValue] {
		ext = ".json"
	} else {
		if  requestedTypes[context.ContentYAMLHeaderValue] {
			ext = ".yml"
		}
	}
	if ext != ".yml" && ext != ".json" {
		log.Errorf("Invalid request type:%s", ext)
		badRequest(ctx, "invalid request type, only json,yaml are supported")
		return
	}
	finalConfig, err := smartConfigFileMerger.Merge(s.repo, appName, appVersion, []string{profile})
	if err != nil {
		log.Errorf("error merging the configuration:%s", err)
		internalServerError(ctx)
		return
	}
	respondConfig(ctx, finalConfig, ext, log)
}
