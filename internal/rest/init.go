package rest

import (
	"context"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/n3wtron/vconf/v2/pkg/configrepo"
	"github.com/sirupsen/logrus"
	"mime"
)

type Server struct {
	repo    configrepo.Repo
	app     *iris.Application
	address string
}

func New(repo configrepo.Repo, address string) *Server {
	s := &Server{repo: repo, address: address}
	app := iris.New()
	app.Logger().SetLevel(logrus.GetLevel().String())
	app.Use(logger.New())
	s.app = app
	s.initV1Api()
	return s
}

func (s *Server) Start() error {
	return s.app.Run(iris.Addr(s.address), iris.WithoutServerError(iris.ErrServerClosed))
}

func (s *Server) Stop() {
	if s.app != nil {
		logrus.Infof("Stopping REST server :%s", s.address)
		err := s.app.Shutdown(context.Background())
		if err != nil {
			logrus.Warnf("Error stopping REST server:%s", err)
		}
	}
}

func (s *Server) initV1Api() {
	v1Api := s.app.Party("/v1")
	s.registerConfigApi(v1Api)
}


func init(){
	initExtraMimeTypes()
}

func initExtraMimeTypes() {
	mime.AddExtensionType(".yml","application/x-yaml")
	mime.AddExtensionType(".yaml","application/x-yaml")
}