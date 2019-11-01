package rest

import (
	"github.com/kataras/iris"
	"net/http"
	"strings"
)

func internalServerError(ctx iris.Context) {
	ctx.StatusCode(http.StatusInternalServerError)
}

func badRequest(ctx iris.Context, err string) {
	ctx.StatusCode(http.StatusBadRequest)
	if err != "" {
		ctx.WriteString(err)
	}
}
func serviceUnavailableResponse(ctx iris.Context) {
	ctx.StatusCode(http.StatusServiceUnavailable)
}

func notFoundResponse(ctx iris.Context) {
	ctx.StatusCode(http.StatusNotFound)
}

func unAuthorizedResponse(ctx iris.Context) {
	ctx.StatusCode(http.StatusUnauthorized)
}

func forbiddenResponse(ctx iris.Context) {
	ctx.StatusCode(http.StatusForbidden)
}

func noContentResponse(ctx iris.Context) {
	ctx.StatusCode(http.StatusNoContent)
}

func getAccepts(ctx iris.Context) map[string]bool {
	result := make(map[string]bool)
	for _, accept := range strings.Split(ctx.GetHeader("Accept"), ",") {
		result[accept] = true
	}
	return result
}
