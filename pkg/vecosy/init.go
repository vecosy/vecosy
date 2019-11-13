package vecosy

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	vecosyGrpc "github.com/vecosy/vecosy/v2/internal/grpc"
	"google.golang.org/grpc"
	"strings"
	"sync"
	"time"
)

type Client struct {
	AppName           string
	AppVersion        string
	Environment       string
	conn              *grpc.ClientConn
	watchClient       vecosyGrpc.WatchServiceClient
	smartConfigClient vecosyGrpc.SmartConfigClient
	viper             *viper.Viper
	updateMutex       sync.Mutex
}

func New(vecosyServer, appName, appVersion, environment string, conf *viper.Viper) (*Client, error) {
	var err error
	vecosyCl := &Client{AppName: appName, AppVersion: appVersion, Environment: environment}
	vecosyCl.initViper(conf)
	vecosyCl.conn, err = grpc.Dial(vecosyServer, grpc.WithInsecure(), grpc.WithBackoffConfig(grpc.BackoffConfig{MaxDelay: 30 * time.Second}))
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

func (vc *Client) initViper(conf *viper.Viper) {
	viperInstance := conf
	if viperInstance == nil {
		viperInstance = viper.GetViper()
	}
	viperInstance.SetConfigType("yaml")
	vc.viper = viperInstance
}

func (vc *Client) WatchChanges() error {
	request := &vecosyGrpc.WatchRequest{
		WatcherName: fmt.Sprintf("%s-watcher", vc.AppName),
		Application: &vecosyGrpc.Application{
			AppName:    vc.AppName,
			AppVersion: vc.AppVersion,
		},
	}
	watchStream, err := vc.watchClient.Watch(context.Background(), request)
	if err != nil {
		return err
	}
	go vc.watchChanges(watchStream)
	return nil
}

func (vc *Client) watchChanges(watcher vecosyGrpc.WatchService_WatchClient) {
	for {
		changes, err := watcher.Recv()
		if err != nil {
			errorDelay := 10 * time.Second
			logrus.Errorf("error watching changes wait %s sec error:%s", errorDelay, err)
			time.Sleep(errorDelay)
			break
		} else {
			if changes.Changed {
				err = vc.UpdateConfig()
				if err != nil {
					logrus.Errorf("Error updating configuration :%s", err)
				}
			}
		}
	}
	watcher.CloseSend()
	// rescheduling myself
	vc.WatchChanges()
}

func (vc *Client) UpdateConfig() error {
	vc.updateMutex.Lock()
	defer vc.updateMutex.Unlock()
	request := &vecosyGrpc.GetConfigRequest{
		AppName:     vc.AppName,
		AppVersion:  vc.AppVersion,
		Environment: vc.Environment,
	}
	response, err := vc.smartConfigClient.GetConfig(context.Background(), request)
	if err != nil {
		logrus.Errorf("Error getting configuration:%s", err)
		return err
	}
	logrus.Debugf("Received %s", response.ConfigContent)
	configReader := strings.NewReader(response.ConfigContent)
	return vc.viper.ReadConfig(configReader)
}
