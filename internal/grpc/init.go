package vconf

import (
	"context"
	"github.com/n3wtron/vconf/v2/pkg/configrepo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	repo    configrepo.Repo
	server  *grpc.Server
	address string
}

func NewServer(repo configrepo.Repo) *Server {
	return &Server{repo: repo}
}

func (s *Server) Start(address string) error {
	s.server = grpc.NewServer()
	s.address = address
	RegisterConfigurationServer(s.server, s)
	logrus.Infof("Starting grpc server on address %s", address)
	listener, err := net.Listen("tcp4", address)
	if err != nil {
		logrus.Errorf("Error creating grpc listener: %s", err)
		return err
	}
	return s.server.Serve(listener)
}

func (s *Server) Stop() {
	logrus.Infof("grpc server with address:%s stopped", s.address)
	s.server.Stop()
}

func (s *Server) GetFile(ctx context.Context, request *GetFileRequest) (*GetFileResponse, error) {
	flContent, err := s.repo.GetFile(request.AppName, request.AppVersion, request.FilePath)
	if err != nil {
		return nil, err
	}
	return &GetFileResponse{
		FileContent: flContent,
	}, nil
}
