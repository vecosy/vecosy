package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/vecosy/vecosy/v2/internal/grpc"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

func startGRPC(repo configrepo.Repo)  {
	viper.SetDefault("server.grpc.address",":8081")
	server := grpc.New(repo, viper.GetString("server.grpc.address"))
	err := server.Start()
	if err != nil {
		logrus.Fatalf("Error starting GPRC server:%s", err)
	}
}
