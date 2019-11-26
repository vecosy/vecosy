package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/vecosy/vecosy/v2/internal/restapi"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

func startRest(cfgRepo configrepo.Repo) {
	viper.SetDefault("server.rest.address", ":8080")
	restSrv := restapi.New(cfgRepo, viper.GetString("server.rest.address"), viper.GetBool("security.enabled"))
	err := restSrv.Start()
	if err != nil {
		logrus.Fatalf("error starting rest server:%s", err)
	}
}
