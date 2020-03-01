package restapi

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestServer_GetSmartConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepo(ctrl)

	appVersion := "v1.0.0"
	appName := "app1"
	app := configrepo.NewApplicationVersion(appName, appVersion)
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

	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	for _, security := range []bool{false, true} {
		srv := New(repo, "127.0.0.1:8080", security)
		ht := httptest.New(t, srv.app)

		repo.EXPECT().GetFile(app, "config.yml").Times(2).Return(&configrepo.RepoFile{
			Version: commitVersion,
			Content: app1CommonYmlContent,
		}, nil)

		repo.EXPECT().GetFile(app, "dev/config.yml").Times(2).Return(&configrepo.RepoFile{
			Version: commitVersion,
			Content: app1DevYmlContent,
		}, nil)

		mergedConfig := map[string]interface{}{"prop1": app1DevContentPop1, "commonProp": app1CommonContentCommonProp, "common": map[string]interface{}{"subProp": app1DevContentCommonSubProp, "subProp2": app1CommonContentSubProp2}}
		t.Run(fmt.Sprintf("getSmartConfig_JSON_security_%v", security), func(t *testing.T) {
			req := ht.GET(fmt.Sprintf("/v1/config/%s/%s/%s", appName, appVersion, "dev"))
			req = req.WithHeader("Accept", context.ContentJSONHeaderValue)
			if security {
				applySecurity(t, privKey, req, repo, app)
			}
			res := req.Expect()
			res.JSON().Equal(mergedConfig)
		})

		t.Run(fmt.Sprintf("getSmartConfig_YAML_security_%v", security), func(t *testing.T) {
			req := ht.GET(fmt.Sprintf("/v1/config/%s/%s/%s", appName, appVersion, "dev"))
			req = req.WithHeader("Accept", context.ContentYAMLHeaderValue)
			if security {
				applySecurity(t, privKey, req, repo, app)
			}
			res := req.Expect()
			rawYml := res.Body().Raw()
			receivedYml := make(map[interface{}]interface{})
			assert.NoError(t, yaml.Unmarshal([]byte(rawYml), &receivedYml))
			normalizedYml, err := utils.NormalizeMap(receivedYml)
			assert.NoError(t, err)
			assert.Equal(t, mergedConfig, normalizedYml)
		})

		if security {
			t.Run("unauthorized", func(t *testing.T) {
				req := ht.GET(fmt.Sprintf("/v1/config/%s/%s/%s", appName, appVersion, "dev"))
				req = req.WithHeader("Accept", context.ContentYAMLHeaderValue)
				repo.EXPECT().GetFile(app, "pub.key").Return(&configrepo.RepoFile{
					Version: uuid.New().String(),
					Content: utils.PublicKeyToBytes(&privKey.PublicKey),
				}, nil)
				res := req.Expect()
				res.Status(httptest.StatusUnauthorized)
			})
		}
	}
}
