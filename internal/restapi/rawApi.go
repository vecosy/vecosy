package restapi

import (
	"github.com/h2non/filetype"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"mime"
	"path/filepath"
)

func (s *Server) registerRawEndpoints(parent iris.Party) {
	configAPI := parent.Party("/raw")
	configAPI.Get("/{appName:string}/{appVersion:string}/{filePath:path}", s.getFile)
}

func (s *Server) getFile(ctx iris.Context) {
	appName := ctx.Params().Get("appName")
	appVersion := ctx.Params().Get("appVersion")
	filePath := ctx.Params().Get("filePath")
	log := logrus.WithField("appName", appName).WithField("appVersion", appVersion).WithField("filePath", filePath)
	log.Infof("GetFile")

	app := configrepo.NewApplicationVersion(appName, appVersion)
	err := checkApplication(ctx, app, log)
	if err != nil {
		return
	}

	err = s.CheckToken(ctx, app)
	if err != nil {
		return
	}

	file, err := s.repo.GetFile(app, filePath)
	if err != nil {
		log.Errorf("error getting file err:%s", err)
		if err == configrepo.ErrFileNotFound {
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
