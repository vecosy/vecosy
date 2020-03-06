package restapi

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/testutil"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestServer_GetFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	appVersion := "v1.0.0"
	appName := "app1"
	app := configrepo.NewApplicationVersion(appName, appVersion)

	commitVersion := uuid.New().String()
	app1DevContent := map[string]interface{}{"prop1": uuid.New().String(), "sub": map[string]interface{}{"subProp1": uuid.New().String()}}
	ymlContent, err := yaml.Marshal(app1DevContent)
	assert.NoError(t, err)

	privKey, _, err := testutil.GenerateKeyPair()
	assert.NoError(t, err)

	for _, security := range []bool{false, true} {
		repo := mocks.NewMockRepo(ctrl)
		srv := New(repo, "127.0.0.1:8080", security)
		ht := httptest.New(t, srv.app)

		t.Run(fmt.Sprintf("valid config security_%v", security), func(t *testing.T) {
			filePath := "config.yml"
			req := ht.GET("/v1/raw/{appName}/{appVersion}/{filePath}")
			req = req.WithPath("appName", appName)
			req = req.WithPath("appVersion", appVersion)
			if security {
				applySecurity(t, privKey, req, repo, app)
			}

			fileReq := req.WithPath("filePath", filePath)
			repo.EXPECT().GetFile(app, filePath).Return(&configrepo.RepoFile{
				Version: commitVersion,
				Content: ymlContent,
			}, nil)

			res := fileReq.Expect()
			rawYml := res.Body().Raw()
			receivedYml := make(map[interface{}]interface{})
			assert.NoError(t, yaml.Unmarshal([]byte(rawYml), &receivedYml))
			normalizedYml, err := utils.NormalizeMap(receivedYml)
			assert.NoError(t, err)
			assert.Equal(t, app1DevContent, normalizedYml)
		})

		t.Run(fmt.Sprintf("not found security_%v", security), func(t *testing.T) {
			filePath := "notExistFile"
			req := ht.GET("/v1/raw/{appName}/{appVersion}/{filePath}")
			req = req.WithPath("appName", appName)
			req = req.WithPath("appVersion", appVersion)
			fileReq := req.WithPath("filePath", filePath)
			if security {
				applySecurity(t, privKey, req, repo, app)
			}
			repo.EXPECT().GetFile(app, filePath).Return(nil, configrepo.ErrFileNotFound)
			fileReq.Expect().Status(httptest.StatusNotFound)
		})

		if security {
			t.Run("unauthorized", func(t *testing.T) {
				filePath := "notExistFile"
				req := ht.GET("/v1/raw/{appName}/{appVersion}/{filePath}")
				req = req.WithPath("appName", appName)
				req = req.WithPath("appVersion", appVersion)
				fileReq := req.WithPath("filePath", filePath)
				repo.EXPECT().GetFile(app, "pub.key").Return(&configrepo.RepoFile{
					Version: uuid.New().String(),
					Content: testutil.PublicKeyToBytes(&privKey.PublicKey),
				}, nil)
				fileReq.Expect().Status(httptest.StatusUnauthorized)
			})
		}
	}
}
