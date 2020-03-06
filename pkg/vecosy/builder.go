package vecosy

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	vecosyGrpc "github.com/vecosy/vecosy/v2/internal/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials"
)

// ClientBuilder represent a client initialization builder
type ClientBuilder struct {
	vecosyServer         string
	appName              string
	appVersion           string
	environment          string
	jwsToken             string
	insecure             bool
	tls                  bool
	certFile             string
	serverDomainOverride string
}

// NewClientBuilder create a new ClientBuilder instance
func NewClientBuilder(vecosyServer, appName, appVersion, environment string) *ClientBuilder {
	return &ClientBuilder{
		vecosyServer: vecosyServer,
		appName:      appName,
		appVersion:   appVersion,
		environment:  environment,
	}
}

// WithJWSToken enables the JWS authentication
func (b *ClientBuilder) WithJWSToken(jwsToken string) *ClientBuilder {
	b.insecure = false
	b.jwsToken = jwsToken
	return b
}

// Insecure disable the JWS authentication
func (b *ClientBuilder) Insecure() *ClientBuilder {
	b.insecure = true
	b.jwsToken = ""
	return b
}

// WithTLS enable the TLS
func (b *ClientBuilder) WithTLS(certFile string) *ClientBuilder {
	b.tls = true
	b.certFile = certFile
	return b
}

// WithDomainOverride TEST ONLY: override the TLS server domain validation
func (b *ClientBuilder) WithDomainOverride(serverDomainOverride string) *ClientBuilder {
	b.serverDomainOverride = serverDomainOverride
	return b
}

// Build will generate a new vecosy client configuration
func (b *ClientBuilder) Build(conf *viper.Viper) (*Client, error) {
	var err error
	vecosyCl := &Client{AppName: b.appName, AppVersion: b.appVersion, Environment: b.environment, jwsToken: b.jwsToken, onChangeHandlers: make([]OnChangeHandler, 0)}
	vecosyCl.initViper(conf)
	var transportOption grpc.DialOption
	if b.tls {
		if b.certFile == "" {
			return nil, errors.New("invalid certfile, did you forgot to call WithTLS method")
		}
		tlsCreds, err := credentials.NewClientTLSFromFile(b.certFile, b.serverDomainOverride)
		if err != nil {
			return nil, err
		}
		transportOption = grpc.WithTransportCredentials(tlsCreds)
	} else {
		transportOption = grpc.WithInsecure()
	}

	vecosyCl.conn, err = grpc.Dial(b.vecosyServer, transportOption, grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoff.DefaultConfig,
		MinConnectTimeout: 0,
	}))
	if err != nil {
		logrus.Errorf("Error dialing grpc:%s", err)
		return nil, err
	}
	vecosyCl.watchClient = vecosyGrpc.NewWatchServiceClient(vecosyCl.conn)
	vecosyCl.smartConfigClient = vecosyGrpc.NewSmartConfigClient(vecosyCl.conn)
	err = vecosyCl.UpdateConfig()
	if err != nil {
		logrus.Errorf("Error updating configuration:%s", err)
		return nil, err
	}
	return vecosyCl, nil
}
