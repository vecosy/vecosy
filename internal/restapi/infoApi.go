package restapi

import (
	"github.com/hashicorp/go-version"
	"github.com/kataras/iris/v12"
)

func (s *Server) registerInfoEndpoints(parent iris.Party) {
	configAPI := parent.Party("/info")
	configAPI.Get("/", s.info)
	configAPI.Get("/{appName:string}/", s.getApp)
}

func versionsToList(versions []*version.Version) []string {
	result := make([]string, len(versions))
	for i, ver := range versions {
		result[i] = ver.String()
	}
	return result
}

func (s *Server) info(ctx iris.Context) {
	appsVersions := s.repo.GetAppsVersions()
	versions := make(map[string][]string)
	for appName, appVer := range appsVersions {
		versions[appName] = versionsToList(appVer)
	}
	_, _ = ctx.JSON(versions)
}

func (s *Server) getApp(ctx iris.Context) {
	appName := ctx.Params().Get("appName")
	appVersions := s.repo.GetAppsVersions()[appName]
	_, _ = ctx.JSON(versionsToList(appVersions))
}
