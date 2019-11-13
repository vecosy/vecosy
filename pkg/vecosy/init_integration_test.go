// +build integration

package vecosy

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/phayes/freeport"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/grpc"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"testing"
	"time"
)

func TestNew_IT(t *testing.T) {
	check := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepo(ctrl)

	appName := "app1"
	appVersion := "1.0.0"
	propValue := uuid.New().String()
	environment := "dev"
	configContent := fmt.Sprintf(`environment: %s
prop: %s`, environment, propValue)
	devConfigFile := &configrepo.RepoFile{
		Version: uuid.New().String(),
		Content: []byte(configContent),
	}
	mockRepo.EXPECT().GetFile(appName, appVersion, "config.yml").Return(nil, fmt.Errorf("file not found"))
	mockRepo.EXPECT().GetFile(appName, appVersion, "dev/config.yml").Return(devConfigFile, nil)

	freePort, err := freeport.GetFreePort()
	assert.NoError(t, err)
	address := fmt.Sprintf("127.0.0.1:%d", freePort)
	srv := grpc.New(mockRepo, address)
	go func() {
		err := srv.Start()
		if err != nil {
			assert.FailNow(t, "error starting grpc server %s", err)
		}
	}()
	time.Sleep(1 * time.Second)
	defer srv.Stop()

	cfg := viper.New()
	cl, err := New(address, appName, appVersion, environment, cfg)
	check.NoError(err)
	check.NotNil(cl)
	check.Equal(cfg.GetString("environment"), "dev")
	check.Equal(cfg.GetString("prop"), propValue)

}
