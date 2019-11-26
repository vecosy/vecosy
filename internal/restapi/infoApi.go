package restapi

import (
	"github.com/hashicorp/go-version"
	"github.com/kataras/iris"
)

func (s *Server) registerInfoEndpoints(parent iris.Party) {
	configApi := parent.Party("/info")
	configApi.Get("/", s.Info)
	configApi.Get("/{appName:string}/", s.GetApp)
}

func versionsToList(versions []*version.Version) []string {
	result := make([]string, len(versions), len(versions))
	for i, ver := range versions {
		result[i] = ver.String()
	}
	return result
}

func (s *Server) Info(ctx iris.Context) {
	appsVersions := s.repo.GetAppsVersions()
	versions := make(map[string][]string)
	for appName, appVer := range appsVersions {
		versions[appName] = versionsToList(appVer)
	}
	_, _ = ctx.JSON(versions)
}

func (s *Server) GetApp(ctx iris.Context) {
	appName := ctx.Params().Get("appName")
	appVersions := s.repo.GetAppsVersions()[appName]
	_, _ = ctx.JSON(versionsToList(appVersions))
}
