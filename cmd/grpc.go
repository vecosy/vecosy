package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/grpc"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
)

func startGRPC(repo configrepo.Repo)  {
	server := grpc.New(repo, ":8081")
	err := server.Start()
	if err != nil {
		logrus.Fatalf("Error starting GPRC server:%s", err)
	}
}
