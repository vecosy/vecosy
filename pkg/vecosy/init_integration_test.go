// +build integration

package vecosy

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/phayes/freeport"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/grpcapi"
	"github.com/vecosy/vecosy/v2/internal/security"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"testing"
	"time"
)

func Test_Client_IT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepo(ctrl)

	appName := "app1"
	appVersion := "1.0.0"
	app := configrepo.NewApplicationVersion(appName, appVersion)
	propValue := uuid.New().String()
	environment := "dev"
	configContent := fmt.Sprintf(`environment: %s
prop: %s`, environment, propValue)
	devConfigFile := &configrepo.RepoFile{
		Version: uuid.New().String(),
		Content: []byte(configContent),
	}
	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)
	for _, security := range []bool{false, true} {
		t.Run(fmt.Sprintf("Client_security_%v", security), func(t *testing.T) {
			check := assert.New(t)
			mockRepo.EXPECT().GetFile(app, "config.yml").Return(nil, fmt.Errorf("file not found"))
			mockRepo.EXPECT().GetFile(app, "dev/config.yml").Return(devConfigFile, nil)
			freePort, err := freeport.GetFreePort()
			assert.NoError(t, err)
			address := fmt.Sprintf("127.0.0.1:%d", freePort)
			srv := grpcapi.New(mockRepo, address, security)
			go func() {
				err := srv.Start()
				if err != nil {
					assert.FailNow(t, "error starting grpc server %s", err)
				}
			}()
			time.Sleep(1 * time.Second)
			defer srv.Stop()

			cfg := viper.New()
			var cl *Client
			if security {
				mockRepo.EXPECT().GetFile(app, "pub.key").Return(&configrepo.RepoFile{
					Version: uuid.New().String(),
					Content: utils.PublicKeyToBytes(&privKey.PublicKey),
				}, nil)
				jws := utils.GenJwsFromPrivateKey(t, privKey, "testApp")
				cl, err = New(address, appName, appVersion, environment, jws.FullSerialize(), cfg)
			} else {
				cl, err = NewInsecure(address, appName, appVersion, environment, cfg)
			}
			check.NoError(err)
			check.NotNil(cl)
			check.Equal(cfg.GetString("environment"), "dev")
			check.Equal(cfg.GetString("prop"), propValue)
		})
	}
}

func Test_Client_Unauthorized_IT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mocks.NewMockRepo(ctrl)

	appName := "app1"
	appVersion := "1.0.0"
	app := configrepo.NewApplicationVersion(appName, appVersion)
	environment := "dev"

	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)
	wrongPrivKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)
	check := assert.New(t)
	freePort, err := freeport.GetFreePort()
	assert.NoError(t, err)
	address := fmt.Sprintf("127.0.0.1:%d", freePort)
	srv := grpcapi.New(mockRepo, address, true)
	go func() {
		err := srv.Start()
		if err != nil {
			assert.FailNow(t, "error starting grpc server %s", err)
		}
	}()
	time.Sleep(1 * time.Second)
	defer srv.Stop()

	cfg := viper.New()
	mockRepo.EXPECT().GetFile(app, "pub.key").Return(&configrepo.RepoFile{
		Version: uuid.New().String(),
		Content: utils.PublicKeyToBytes(&privKey.PublicKey),
	}, nil)
	jws := utils.GenJwsFromPrivateKey(t, wrongPrivKey, "testApp")
	cl, err := New(address, appName, appVersion, environment, jws.FullSerialize(), cfg)
	check.Error(err)
	check.Contains(err.Error(), security.AuthFailed.Error())
	check.Nil(cl)
}
