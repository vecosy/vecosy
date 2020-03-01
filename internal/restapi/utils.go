package restapi

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/validation"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"net/http"
	"strings"
)

func internalServerError(ctx iris.Context) {
	ctx.StatusCode(http.StatusInternalServerError)
}

func badRequest(ctx iris.Context, err string) {
	ctx.StatusCode(http.StatusBadRequest)
	if err != "" {
		_, _ = ctx.WriteString(err)
	}
}

func notFoundResponse(ctx iris.Context) {
	ctx.StatusCode(http.StatusNotFound)
}

func unAuthorizedResponse(ctx iris.Context) {
	ctx.StatusCode(http.StatusUnauthorized)
}

func getAccepts(ctx iris.Context) map[string]bool {
	result := make(map[string]bool)
	for _, accept := range strings.Split(ctx.GetHeader("Accept"), ",") {
		result[accept] = true
	}
	return result
}

func checkApplication(ctx iris.Context, app *configrepo.ApplicationVersion, log *logrus.Entry) error {
	err := validation.ValidateApplicationVersion(app)
	if err != nil {
		log.Errorf("invalid application: %s", err)
		badRequest(ctx, fmt.Sprintf("%+v is not a application", app))
	}
	return err
}
