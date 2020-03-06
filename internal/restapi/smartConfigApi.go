package restapi

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/merger"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

var smartConfigFileMerger = merger.SmartConfigMerger{}

func (s *Server) registerSmartConfigEndpoints(parent iris.Party) {
	configAPI := parent.Party("/config")
	configAPI.Get("/", s.info)
	configAPI.Get("/{appName:string}/{appVersion:string}/{profile:string}", s.getSmartConfig)
}

func (s *Server) getSmartConfig(ctx iris.Context) {
	appName := ctx.Params().GetString("appName")
	appVersion := ctx.Params().GetString("appVersion")
	profile := ctx.Params().GetString("profile")
	requestedTypes := getAccepts(ctx)
	log := logrus.WithField("appName", appName).WithField("appVersion", appVersion).WithField("profile", profile)
	log = log.WithField("requested types", requestedTypes)
	log.Info("GetSmartConfig")

	app := configrepo.NewApplicationVersion(appName, appVersion)
	err := checkApplication(ctx, app, log)
	if err != nil {
		return
	}

	err = s.CheckToken(ctx, app)
	if err != nil {
		return
	}

	var ext string
	if requestedTypes[context.ContentJSONHeaderValue] || requestedTypes[context.ContentHTMLHeaderValue] {
		ext = ".json"
	} else {
		if requestedTypes[context.ContentYAMLHeaderValue] {
			ext = ".yml"
		}
	}
	if ext != ".yml" && ext != ".json" {
		log.Errorf("Invalid request type:%s", ext)
		badRequest(ctx, "invalid request type, only json,yaml are supported")
		return
	}
	finalConfig, err := smartConfigFileMerger.Merge(s.repo, app, []string{profile})
	if err != nil {
		log.Errorf("error merging the configuration:%s", err)
		internalServerError(ctx)
		return
	}
	respondConfig(ctx, finalConfig, ext, log)
}
