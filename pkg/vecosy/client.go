package vecosy

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	vecosyGrpc "github.com/vecosy/vecosy/v2/internal/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
	"sync"
	"time"
)

type OnChangeHandler = func()

type Client struct {
	AppName           string
	AppVersion        string
	Environment       string
	jwsToken          string
	conn              *grpc.ClientConn
	watchClient       vecosyGrpc.WatchServiceClient
	smartConfigClient vecosyGrpc.SmartConfigClient
	viper             *viper.Viper
	updateMutex       sync.Mutex
	onChangeHandlers  []OnChangeHandler
}

func (vc *Client) initViper(conf *viper.Viper) {
	viperInstance := conf
	if viperInstance == nil {
		viperInstance = viper.GetViper()
	}
	viperInstance.SetConfigType("yaml")
	vc.viper = viperInstance
}

func (vc *Client) genContext(parent context.Context) context.Context {
	if vc.jwsToken != "" {
		return metadata.AppendToOutgoingContext(parent, "token", vc.jwsToken)
	}
	return parent
}

func (vc *Client) WatchChanges() error {
	request := &vecosyGrpc.WatchRequest{
		WatcherName: fmt.Sprintf("%s-watcher", vc.AppName),
		Application: &vecosyGrpc.Application{
			AppName:    vc.AppName,
			AppVersion: vc.AppVersion,
		},
	}
	watchStream, err := vc.watchClient.Watch(vc.genContext(context.Background()), request)
	if err != nil {
		return err
	}
	go vc.watchChanges(watchStream)
	return nil
}

func (vc *Client) AddOnChangeHandler(handler OnChangeHandler) {
	vc.onChangeHandlers = append(vc.onChangeHandlers, handler)
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
				for _, onChangeHandler := range vc.onChangeHandlers {
					onChangeHandler()
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
	response, err := vc.smartConfigClient.GetConfig(vc.genContext(context.Background()), request)
	if err != nil {
		logrus.Errorf("Error getting configuration:%s", err)
		return err
	}
	logrus.Debugf("Received %s", response.ConfigContent)
	configReader := strings.NewReader(response.ConfigContent)
	return vc.viper.ReadConfig(configReader)
}
