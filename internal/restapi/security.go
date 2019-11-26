package restapi

import (
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/security"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

func (s *Server) CheckToken(ctx iris.Context, app *configrepo.ApplicationVersion) error {
	log := logrus.WithField("method", "CheckToken")
	if !s.IsSecurityEnabled() {
		return nil
	}
	token := ctx.GetHeader("token")
	log.Debugf("checking token:%s", token)
	err := security.CheckJwtToken(s.repo, app, token)
	if err != nil {
		unAuthorizedResponse(ctx)
		return err
	}
	return nil
}
