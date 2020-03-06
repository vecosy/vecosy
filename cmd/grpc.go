package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/vecosy/vecosy/v2/internal/grpcapi"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

func startGRPC(repo configrepo.Repo) {
	var err error
	viper.SetDefault("server.grpc.address", ":8081")
	var server *grpcapi.Server
	if viper.GetBool("server.tls.enabled") {
		server, err = grpcapi.NewTLS(repo, viper.GetString("server.grpc.address"), viper.GetBool("security.enabled"), viper.GetString("server.tls.certificateFile"), viper.GetString("server.tls.keyFile"))
	} else {
		server, err = grpcapi.NewNoTLS(repo, viper.GetString("server.grpc.address"), viper.GetBool("security.enabled"))
	}
	if err != nil {
		logrus.Fatalf("Error starting GPRC server:%s", err)
	}
	err = server.Start()
	if err != nil {
		logrus.Fatalf("Error starting GPRC server:%s", err)
	}
}
