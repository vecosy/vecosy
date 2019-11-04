package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/rest"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

func startRest(cfgRepo configrepo.Repo) {
	restSrv := rest.New(cfgRepo, ":8080")
	err := restSrv.Start()
	if err != nil {
		logrus.Fatalf("error starting rest server:%s", err)
	}
}