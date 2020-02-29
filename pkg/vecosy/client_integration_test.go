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
	"os"
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
	jwsPrivKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	certFile, certKeyFile, err := utils.GenerateCertificateFiles()
	assert.NoError(t, err)
	defer os.Remove(certFile)
	defer os.Remove(certKeyFile)

	for _, tlsEnabled := range []bool{false, true} {
		for _, secEnabled := range []bool{false, true} {
			t.Run(fmt.Sprintf("Client_security_%v_tls_%v", secEnabled, tlsEnabled), func(t *testing.T) {
				check := assert.New(t)
				mockRepo.EXPECT().GetFile(app, "config.yml").Return(nil, fmt.Errorf("file not found"))
				mockRepo.EXPECT().GetFile(app, "dev/config.yml").Return(devConfigFile, nil)
				freePort, err := freeport.GetFreePort()
				assert.NoError(t, err)
				address := fmt.Sprintf("127.0.0.1:%d", freePort)
				var srv *grpcapi.Server
				if tlsEnabled {
					srv, err = grpcapi.NewTLS(mockRepo, address, secEnabled, certFile, certKeyFile)
				} else {
					srv, err = grpcapi.NewNoTLS(mockRepo, address, secEnabled)
				}
				check.NoError(err)
				go func() {
					err := srv.Start()
					if err != nil {
						assert.FailNow(t, "error starting grpc server %s", err)
					}
				}()
				time.Sleep(1 * time.Second)
				defer srv.Stop()

				cfg := viper.New()
				builder := NewBuilder(address, appName, appVersion, environment)
				if secEnabled {
					mockRepo.EXPECT().GetFile(app, "pub.key").Return(&configrepo.RepoFile{
						Version: uuid.New().String(),
						Content: utils.PublicKeyToBytes(&jwsPrivKey.PublicKey),
					}, nil)
					jws := utils.GenJwsFromPrivateKey(t, jwsPrivKey, "testApp")
					builder.WithJWSToken(jws.FullSerialize())
				} else {
					builder.Insecure()
				}
				if tlsEnabled {
					builder.WithTLS(certFile)
				}
				cl, err := builder.Build(cfg)
				check.NoError(err)
				check.NotNil(cl)
				check.Equal(cfg.GetString("environment"), "dev")
				check.Equal(cfg.GetString("prop"), propValue)
			})
		}
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
	srv, err := grpcapi.NewNoTLS(mockRepo, address, true)
	check.NoError(err)
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
	cl, err := NewBuilder(address, appName, appVersion, environment).
		WithJWSToken(jws.FullSerialize()).
		Build(cfg)
	check.Error(err)
	check.Contains(err.Error(), security.AuthFailed.Error())
	check.Nil(cl)
}
