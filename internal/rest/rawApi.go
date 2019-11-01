package rest

import (
	"fmt"
	"github.com/h2non/filetype"
	"github.com/hashicorp/go-version"
	"github.com/kataras/iris"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"github.com/sirupsen/logrus"
	"mime"
	"path/filepath"
)

func (s *Server) registerFileEndpoints(parent iris.Party) {
	configApi := parent.Party("/raw")
	configApi.Get("/", s.Info)
	configApi.Get("/{appName:string}/", s.GetApp)
	configApi.Get("/{appName:string}/{appVersion:string}/{filePath:path}", s.GetFile)
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

func (s *Server) GetFile(ctx iris.Context) {
	appName := ctx.Params().Get("appName")
	appVersion := ctx.Params().Get("appVersion")
	filePath := ctx.Params().Get("filePath")
	log := logrus.WithField("appName", appName).WithField("appVersion", appVersion).WithField("filePath", filePath)
	log.Infof("GetFile")

	_, err := version.NewVersion(appVersion)
	if err != nil {
		log.Errorf("invalid version: %s", err)
		badRequest(ctx, fmt.Sprintf("%s is not a valid version", appVersion))
	}

	file, err := s.repo.GetFile(appName, appVersion, filePath)
	if err != nil {
		log.Errorf("error getting file err:%s", err)
		if err == configrepo.FileNotFoundError {
			notFoundResponse(ctx)
		} else {
			internalServerError(ctx)
		}
		return
	}
	var mimeType string
	fileKind, err := filetype.Match(file.Content)
	if err == nil && fileKind != filetype.Unknown {
		log.Debugf("Found fileType: %+v", fileKind)
		mimeType = fileKind.MIME.Value
	} else {
		log.Infof("file extension:%s", filepath.Ext(filePath))
		mimeType = mime.TypeByExtension(filepath.Ext(filePath))
	}
	log.Debugf("Detected fileType:%s", mimeType)
	ctx.ContentType(mimeType)
	_, _ = ctx.Write(file.Content)
}
