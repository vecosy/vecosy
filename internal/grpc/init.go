package vconf

import (
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"google.golang.org/grpc"
	"net"
	"sync"
)

type Watcher struct {
	id          string
	watcherName string
	apps        map[string]*version.Version
}

type Server struct {
	repo           configrepo.Repo
	server         *grpc.Server
	address        string
	watchers       sync.Map
	watcherStreams sync.Map
}

func New(repo configrepo.Repo, address string) *Server {
	s := &Server{repo: repo, address: address}
	s.server = grpc.NewServer()
	RegisterRawServer(s.server, s)
	RegisterSmartConfigServer(s.server, s)
	RegisterWatchServiceServer(s.server, s)
	return s
}

func (s *Server) Start() error {
	logrus.Infof("Starting grpc server on address %s", s.address)
	listener, err := net.Listen("tcp4", s.address)
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
