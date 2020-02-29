package restapi

import (
	"context"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"mime"
)

type Server struct {
	repo            configrepo.Repo
	app             *iris.Application
	address         string
	securityEnabled bool
}

func New(repo configrepo.Repo, address string, securityEnabled bool) *Server {
	s := &Server{repo: repo, address: address, securityEnabled: securityEnabled}
	log := logrus.WithField("address", address).WithField("securityEnabled", securityEnabled)
	log.Info("Rest server created")
	app := iris.New()
	app.Logger().SetLevel(logrus.GetLevel().String())
	app.Use(logger.New())
	s.app = app
	s.initAlive()
	s.initV1Api()
	return s
}

func (s *Server) StartTLS(certFile, keyFile string) error {
	logrus.Infof("Start rest with TLS cert:%s key:%s", certFile, keyFile)
	return s.app.Run(iris.TLS(s.address, certFile, keyFile), iris.WithoutServerError(iris.ErrServerClosed))
}

func (s *Server) StartNoTLS() error {
	logrus.Info("Start rest with NO TLS")
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

func (s *Server) initAlive() {
	s.app.Get("/alive", func(ctx iris.Context) {
		_, _ = ctx.WriteString("alive")
	})
}

func (s *Server) initV1Api() {
	v1Api := s.app.Party("/v1")
	s.registerInfoEndpoints(v1Api)
	s.registerRawEndpoints(v1Api)
	s.registerSmartConfigEndpoints(v1Api)
	s.registerSpringCloudEndpoints(v1Api)
}

func (s *Server) IsSecurityEnabled() bool {
	return s.securityEnabled
}

func init() {
	initExtraMimeTypes()
}

func initExtraMimeTypes() {
	mime.AddExtensionType(".yml", "application/x-yaml")
	mime.AddExtensionType(".yaml", "application/x-yaml")
}
