package rest

import (
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/kataras/iris/httptest"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestServer_GetFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepo(ctrl)
	srv := New(repo, "127.0.0.1:8080")
	ht := httptest.New(t, srv.app)

	appVersion := "v1.0.0"
	appName := "app1"
	commitVersion := uuid.New().String()
	app1DevContent := map[string]interface{}{"prop1": uuid.New().String(), "sub": map[string]interface{}{"subProp1": uuid.New().String()}}
	ymlContent, err := yaml.Marshal(app1DevContent)
	assert.NoError(t, err)


	t.Run("valid config", func(t *testing.T) {
		filePath := "config.yml"
		req := ht.GET("/v1/config/{appName}/{appVersion}/{filePath}")
		req = req.WithPath("appName", appName)
		req = req.WithPath("appVersion", appVersion)
		fileReq := req.WithPath("filePath", filePath)
		repo.EXPECT().GetFile(appName, appVersion, filePath).Return(&configrepo.RepoFile{
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

	t.Run("not found", func(t *testing.T) {
		filePath := "notExistFile"
		req := ht.GET("/v1/config/{appName}/{appVersion}/{filePath}")
		req = req.WithPath("appName", appName)
		req = req.WithPath("appVersion", appVersion)
		fileReq := req.WithPath("filePath", filePath)
		repo.EXPECT().GetFile(appName, appVersion, filePath).Return(nil, configrepo.FileNotFoundError)
		fileReq.Expect().Status(httptest.StatusNotFound)
	})
}
