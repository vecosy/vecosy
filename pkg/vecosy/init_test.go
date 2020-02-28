// +build !race

package vecosy

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/grpcapi"
	"io"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	m.Run()
}

func TestClient_UpdateConfig(t *testing.T) {
	checks := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSmartConfigCl := grpcapi.NewMockSmartConfigClient(ctrl)
	appName := "app1"
	appVersion := "1.0.0"
	environment := "dev"
	vecosyCl := &Client{AppName: appName, AppVersion: appVersion, Environment: environment, smartConfigClient: mockSmartConfigCl}
	cfg := viper.New()
	vecosyCl.initViper(cfg)
	checks.NotNil(vecosyCl.viper)
	request := &grpcapi.GetConfigRequest{
		AppName:     appName,
		AppVersion:  appVersion,
		Environment: environment,
	}
	propValue := uuid.New().String()
	configContent := fmt.Sprintf(`environment: %s
prop: %s`, environment, propValue)
	response := &grpcapi.GetConfigResponse{ConfigContent: configContent}
	mockSmartConfigCl.EXPECT().GetConfig(gomock.Any(), request).Return(response, nil)
	checks.NoError(vecosyCl.UpdateConfig())
	checks.Equal(cfg.GetString("environment"), environment)
	checks.Equal(cfg.GetString("prop"), propValue)
}

func TestClient_WatchChanges(t *testing.T) {
	checks := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSmartConfigCl := grpcapi.NewMockSmartConfigClient(ctrl)
	mockWatchCl := grpcapi.NewMockWatchServiceClient(ctrl)
	appName := "app1"
	appVersion := "1.0.0"
	environment := "dev"
	vecosyCl := &Client{AppName: appName, AppVersion: appVersion, Environment: environment, smartConfigClient: mockSmartConfigCl, watchClient: mockWatchCl}
	cfg := viper.New()
	vecosyCl.initViper(cfg)
	checks.NotNil(vecosyCl.viper)
	request := &grpcapi.GetConfigRequest{
		AppName:     appName,
		AppVersion:  appVersion,
		Environment: environment,
	}
	propValue1 := uuid.New().String()
	configContent1 := fmt.Sprintf(`environment: %s
prop: %s`, environment, propValue1)
	response1 := &grpcapi.GetConfigResponse{ConfigContent: configContent1}
	mockSmartConfigCl.EXPECT().GetConfig(gomock.Any(), request).Return(response1, nil)
	checks.NoError(vecosyCl.UpdateConfig())

	watchRequest := &grpcapi.WatchRequest{
		WatcherName: "app1-watcher",
		Application: &grpcapi.Application{
			AppName:    appName,
			AppVersion: appVersion,
		},
	}
	watchResponse := grpcapi.NewMockWatchService_WatchClient(ctrl)
	watchResponse.EXPECT().Recv().Return(&grpcapi.WatchResponse{Changed: true}, nil)
	mockWatchCl.EXPECT().Watch(gomock.Any(), watchRequest).Return(watchResponse, nil)
	watchResponse.EXPECT().Recv().Return(nil, io.EOF)

	propValue2 := uuid.New().String()
	configContent2 := fmt.Sprintf(`environment: %s
prop: %s`, environment, propValue2)
	response2 := &grpcapi.GetConfigResponse{ConfigContent: configContent2}
	mockSmartConfigCl.EXPECT().GetConfig(gomock.Any(), request).Return(response2, nil)

	onChangeFnCalled := false
	onChangeFn := func() { onChangeFnCalled = true }
	vecosyCl.AddOnChangeHandler(onChangeFn)

	checks.NoError(vecosyCl.WatchChanges())
	time.Sleep(2 * time.Second)
	checks.Equal(cfg.GetString("environment"), environment)
	checks.Equal(cfg.GetString("prop"), propValue2)
	checks.True(onChangeFnCalled)
}
