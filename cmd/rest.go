package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/vecosy/vecosy/v2/internal/restapi"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

func startRest(cfgRepo configrepo.Repo) {
	var err error
	viper.SetDefault("server.rest.address", ":443")
	restSrv := restapi.New(cfgRepo, viper.GetString("server.rest.address"), viper.GetBool("security.enabled"))
	if viper.GetBool("server.tls.enabled") {
		err = restSrv.StartTLS(viper.GetString("server.tls.certificateFile"), viper.GetString("server.tls.keyFile"))
	} else {
		err = restSrv.StartNoTLS()
	}
	if err != nil {
		logrus.Fatalf("error starting rest server:%s", err)
	}
}
