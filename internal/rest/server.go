package rest

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	app := iris.New()
	app.Logger().SetLevel(logrus.GetLevel().String())
	app.Use(recover.New())
	app.Use(logger.New())

	err := app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
	if err != nil {
		logrus.Fatal(err)
	}
}
