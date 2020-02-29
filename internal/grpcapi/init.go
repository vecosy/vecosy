package grpcapi

import (
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"sync"
)

type Watcher struct {
	id          string
	watcherName string
	appName     string
	appVersion  *version.Version
	ch          chan *WatchResponse
}

type Server struct {
	repo            configrepo.Repo
	server          *grpc.Server
	address         string
	watchers        sync.Map
	securityEnabled bool
}

func NewTLS(repo configrepo.Repo, address string, securityEnabled bool, certFile, keyFile string) (*Server, error) {
	tlsCreds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	s := &Server{repo: repo, address: address, securityEnabled: securityEnabled}
	s.server = grpc.NewServer(grpc.Creds(tlsCreds))
	s.registerServices()
	return s, nil
}

func NewNoTLS(repo configrepo.Repo, address string, securityEnabled bool) (*Server, error) {
	s := &Server{repo: repo, address: address, securityEnabled: securityEnabled}
	s.server = grpc.NewServer()
	s.registerServices()
	return s, nil
}

func (s *Server) registerServices() {
	RegisterRawServer(s.server, s)
	RegisterSmartConfigServer(s.server, s)
	RegisterWatchServiceServer(s.server, s)
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

func (s *Server) IsSecurityEnabled() bool {
	return s.securityEnabled
}
