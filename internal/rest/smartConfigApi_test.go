package rest

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/kataras/iris/context"
	"github.com/kataras/iris/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestServer_GetSmartConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepo(ctrl)
	srv := New(repo, "127.0.0.1:8080")
	ht := httptest.New(t, srv.app)

	appVersion := "v1.0.0"
	appName := "app1"
	commitVersion := uuid.New().String()
	app1DevContentCommonSubProp := uuid.New().String()
	app1DevContentPop1 := uuid.New().String()
	app1CommonContentCommonProp := uuid.New().String()
	app1CommonContentSubProp2 := uuid.New().String()

	app1DevContent := map[string]interface{}{"prop1": app1DevContentPop1, "common": map[string]interface{}{"subProp": app1DevContentCommonSubProp}}
	app1CommonContent := map[string]interface{}{"commonProp": app1CommonContentCommonProp, "common": map[string]interface{}{"subProp": uuid.New().String(), "subProp2": app1CommonContentSubProp2}}

	app1CommonYmlContent, err := yaml.Marshal(app1CommonContent)
	assert.NoError(t, err)
	app1DevYmlContent, err := yaml.Marshal(app1DevContent)
	assert.NoError(t, err)

	repo.EXPECT().GetFile(appName, appVersion, "config.yml").Times(4).Return(&configrepo.RepoFile{
		Version: commitVersion,
		Content: app1CommonYmlContent,
	}, nil)

	repo.EXPECT().GetFile(appName, appVersion, "dev/config.yml").Times(4).Return(&configrepo.RepoFile{
		Version: commitVersion,
		Content: app1DevYmlContent,
	}, nil)

	mergedConfig := map[string]interface{}{"prop1": app1DevContentPop1, "commonProp": app1CommonContentCommonProp, "common": map[string]interface{}{"subProp": app1DevContentCommonSubProp, "subProp2": app1CommonContentSubProp2}}
	t.Run("getSmartConfig_JSON", func(t *testing.T) {
		req := ht.GET(fmt.Sprintf("/v1/config/%s/%s/%s", appName, appVersion, "dev"))
		req = req.WithHeader("Accept", context.ContentJSONHeaderValue)
		res := req.Expect()
		res.JSON().Equal(mergedConfig)
	})

	t.Run("getSmartConfig_YAML", func(t *testing.T) {
		req := ht.GET(fmt.Sprintf("/v1/config/%s/%s/%s", appName, appVersion, "dev"))
		req = req.WithHeader("Accept", context.ContentYAMLHeaderValue)
		res := req.Expect()
		rawYml := res.Body().Raw()
		receivedYml := make(map[interface{}]interface{})
		assert.NoError(t, yaml.Unmarshal([]byte(rawYml), &receivedYml))
		normalizedYml, err := utils.NormalizeMap(receivedYml)
		assert.NoError(t, err)
		assert.Equal(t, mergedConfig, normalizedYml)
	})

}
