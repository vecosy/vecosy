package vecosy

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	vconf "github.com/vecosy/vecosy/v2/internal/grpc"
	"google.golang.org/grpc"
	"strings"
)

type Client struct {
	AppName           string
	AppVersion        string
	Environment       string
	conn              *grpc.ClientConn
	watchClient       vconf.WatchServiceClient
	smartConfigClient vconf.SmartConfigClient
	viper             *viper.Viper
}

func New(vecosyServer, appName, appVersion, environment string, conf *viper.Viper) (*Client, error) {
	var err error
	viperInstance := conf
	if viperInstance == nil {
		viperInstance = viper.GetViper()
	}
	viperInstance.SetConfigType("yaml")
	vecosyCl := &Client{AppName: appName, AppVersion: appVersion, Environment: environment, viper: viperInstance}
	vecosyCl.conn, err = grpc.Dial(vecosyServer, grpc.WithInsecure())
	if err != nil {
		logrus.Errorf("Error dialing grpc:%s", err)
		return nil, err
	}
	vecosyCl.watchClient = vconf.NewWatchServiceClient(vecosyCl.conn)
	vecosyCl.smartConfigClient = vconf.NewSmartConfigClient(vecosyCl.conn)
	err = vecosyCl.UpdateConfig()
	if err != nil {
		logrus.Errorf("Error updating configuration:%s", err)
		return nil, err
	}
	return vecosyCl, nil
}

func (vc *Client) WatchChanges() error {
	request := &vconf.WatchRequest{
		WatcherName: fmt.Sprintf("%s-watcher", vc.AppName),
		Application: &vconf.Application{
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

func (vc *Client) watchChanges(watcher vconf.WatchService_WatchClient) {
	for {
		changes, err := watcher.Recv()
		if err != nil {
			logrus.Errorf("error watching changes :%s", err)
			return
		}
		if changes.Changed {
			err = vc.UpdateConfig()
			if err != nil {
				logrus.Errorf("Error updating configuration :%s", err)
			}
		}
	}
}

func (vc *Client) UpdateConfig() error {
	request := &vconf.GetConfigRequest{
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
