package restapi

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jeremywohl/flatten"
	"github.com/kataras/iris/httptest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/vecosy/vecosy/v2/internal/utils"
	"github.com/vecosy/vecosy/v2/mocks"
	"github.com/vecosy/vecosy/v2/pkg/configrepo"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	m.Run()
}

func Test_extractAppNameAndVersion(t *testing.T) {
	tests := []struct {
		name          string
		appAndProfile string
		appName       string
		extension     string
		profile       string
	}{
		{"simple", "app-dev.yml", "app", ".yml", "dev"},
		{"simple_noProfile", "app.yml", "app", ".yml", ""},
		{"application_with_minus", "my-application-dev.json", "my-application", ".json", "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAppName, gotExtension, gotProfile := extractAppNameAndVersion(tt.appAndProfile)
			if gotAppName != tt.appName {
				t.Errorf("extractAppNameAndVersion() gotAppName = %v, appName %v", gotAppName, tt.appName)
			}
			if gotExtension != tt.extension {
				t.Errorf("extractAppNameAndVersion() gotExtension = %v, extension %v", gotExtension, tt.extension)
			}
			if gotProfile != tt.profile {
				t.Errorf("extractAppNameAndVersion() gotProfile = %v, profile %v", gotProfile, tt.profile)
			}
		})
	}
}

func TestServer_Get_OnlyProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepo(ctrl)

	appVersion := "v1.0.0"
	appName := "app1"
	app := configrepo.NewApplicationVersion(appName, appVersion)
	profiles := []string{"dev"}
	appProfiles := strings.Join(profiles, ",")
	commitVersion := uuid.New().String()
	app1DevContent := map[string]interface{}{"prop1": uuid.New().String(), "sub": map[string]interface{}{"subProp1": uuid.New().String()}}
	ymlContent, err := yaml.Marshal(app1DevContent)
	assert.NoError(t, err)
	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	for _, security := range []bool{false, true} {
		srv := New(repo, "127.0.0.1:8080", security)
		ht := httptest.New(t, srv.app)
		repo.EXPECT().GetFile(app, "app1.yml").Times(3).Return(nil, fmt.Errorf("not found"))
		repo.EXPECT().GetFile(app, "application.yml").Times(3).Return(nil, fmt.Errorf("not found"))
		repo.EXPECT().GetFile(app, "application-dev.yml").Times(3).Return(nil, fmt.Errorf("not found"))
		repo.EXPECT().GetFile(app, "app1-dev.yml").Times(3).Return(&configrepo.RepoFile{
			Version: commitVersion,
			Content: ymlContent,
		}, nil)

		t.Run(fmt.Sprintf("springAppInfo_security_%v", security), func(t *testing.T) {
			flattenApp1DevContent, err := flatten.Flatten(app1DevContent, "", flatten.DotStyle)
			assert.NoError(t, err)
			expectedPropertySources := []*propertySources{
				{
					Name:    "app1-dev.yml",
					Source:  flattenApp1DevContent,
					version: commitVersion,
				},
			}
			expectedResult := springSummaryResponse{
				Name:            appName,
				Profiles:        profiles,
				Label:           nil,
				Version:         commitVersion,
				State:           nil,
				PropertySources: expectedPropertySources,
			}

			req := ht.GET(fmt.Sprintf("/v1/spring/%s/%s/%s", appVersion, appName, appProfiles))
			if security {
				applySecurity(t, privKey, req, repo, app)
			}
			res := req.Expect()
			res.JSON().Equal(expectedResult)
		})

		t.Run(fmt.Sprintf("getSpringConfigByAppAndProfile_JSON_security_%v", security), func(t *testing.T) {
			req := ht.GET(fmt.Sprintf("/v1/spring/%s/%s-%s.json", appVersion, appName, appProfiles))
			if security {
				applySecurity(t, privKey, req, repo, app)
			}
			res := req.Expect()
			res.JSON().Equal(app1DevContent)
		})

		t.Run(fmt.Sprintf("getSpringConfigByAppAndProfile_YAML_security_%v", security), func(t *testing.T) {
			req := ht.GET(fmt.Sprintf("/v1/spring/%s/%s-%s.yml", appVersion, appName, appProfiles))
			if security {
				applySecurity(t, privKey, req, repo, app)
			}
			res := req.Expect()
			rawYml := res.Body().Raw()
			receivedYml := make(map[string]interface{})
			assert.NoError(t, yaml.Unmarshal([]byte(rawYml), &receivedYml))
			assert.Equal(t, app1DevContent, receivedYml)
		})
		if security {
			t.Run("unauthorized", func(t *testing.T) {
				repo.EXPECT().GetFile(app, "pub.key").Return(&configrepo.RepoFile{
					Version: uuid.New().String(),
					Content: utils.PublicKeyToBytes(&privKey.PublicKey),
				}, nil)
				req := ht.GET(fmt.Sprintf("/v1/spring/%s/%s-%s.yml", appVersion, appName, appProfiles))
				res := req.Expect()
				res.Status(httptest.StatusUnauthorized)
			})
		}
	}
}

func TestServer_Get_ProfileAndCommon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepo(ctrl)

	appVersion := "v1.0.0"
	appName := "app1"
	app := configrepo.NewApplicationVersion(appName, appVersion)
	devProfiles := []string{"dev"}
	devStrProfiles := strings.Join(devProfiles, ",")
	devAndIntProfiles := []string{"dev", "int"}
	devAndIntStrProfiles := strings.Join(devAndIntProfiles, ",")
	commitVersion := uuid.New().String()
	app1DevContentCommonSubProp := uuid.New().String()
	app1DevContentPop1 := uuid.New().String()
	app1CommonContentCommonProp := uuid.New().String()
	app1CommonContentSubProp2 := uuid.New().String()

	app1DevContent := map[string]interface{}{"prop1": app1DevContentPop1, "common": map[string]interface{}{"subProp": app1DevContentCommonSubProp}}
	app1IntContent := map[string]interface{}{"intProp": uuid.New().String(), "common": map[string]interface{}{"subProp": app1DevContentCommonSubProp}}
	app1CommonContent := map[string]interface{}{"commonProp": app1CommonContentCommonProp, "common": map[string]interface{}{"subProp": uuid.New().String(), "subProp2": app1CommonContentSubProp2}}

	app1CommonYmlContent, err := yaml.Marshal(app1CommonContent)
	assert.NoError(t, err)
	app1DevYmlContent, err := yaml.Marshal(app1DevContent)
	assert.NoError(t, err)
	app1IntYmlContent, err := yaml.Marshal(app1IntContent)
	assert.NoError(t, err)
	privKey, _, err := utils.GenerateKeyPair()
	assert.NoError(t, err)

	for _, security := range []bool{false, true} {
		srv := New(repo, "127.0.0.1:8080", security)
		ht := httptest.New(t, srv.app)
		repo.EXPECT().GetFile(app, "application.yml").Times(4).Return(&configrepo.RepoFile{
			Version: commitVersion,
			Content: app1CommonYmlContent,
		}, nil)

		repo.EXPECT().GetFile(app, "app1.yml").Times(4).Return(&configrepo.RepoFile{
			Version: commitVersion,
			Content: app1CommonYmlContent,
		}, nil)

		repo.EXPECT().GetFile(app, "application-dev.yml").Times(4).Return(&configrepo.RepoFile{
			Version: commitVersion,
			Content: app1CommonYmlContent,
		}, nil)

		repo.EXPECT().GetFile(app, "app1-dev.yml").Times(4).Return(&configrepo.RepoFile{
			Version: commitVersion,
			Content: app1DevYmlContent,
		}, nil)

		repo.EXPECT().GetFile(app, "application-int.yml").Times(1).Return(&configrepo.RepoFile{
			Version: commitVersion,
			Content: app1CommonYmlContent,
		}, nil)

		repo.EXPECT().GetFile(app, "app1-int.yml").Return(&configrepo.RepoFile{
			Version: commitVersion,
			Content: app1IntYmlContent,
		}, nil)

		t.Run(fmt.Sprintf("getSpringFileByAppAndProfile_SingleProfile_security:%v", security), func(t *testing.T) {
			flattenApp1DevContent, err := flatten.Flatten(app1DevContent, "", flatten.DotStyle)
			assert.NoError(t, err)
			flattenApp1CommonContent, err := flatten.Flatten(app1CommonContent, "", flatten.DotStyle)
			assert.NoError(t, err)
			expectedPropertySources := []*propertySources{
				{
					Name:    "app1-dev.yml",
					Source:  flattenApp1DevContent,
					version: commitVersion,
				},
				{
					Name:    "application-dev.yml",
					Source:  flattenApp1CommonContent,
					version: commitVersion,
				},
				{
					Name:    "app1.yml",
					Source:  flattenApp1CommonContent,
					version: commitVersion,
				},
				{
					Name:    "application.yml",
					Source:  flattenApp1CommonContent,
					version: commitVersion,
				},
			}
			expectedResult := springSummaryResponse{
				Name:            appName,
				Profiles:        devProfiles,
				Label:           nil,
				Version:         commitVersion,
				State:           nil,
				PropertySources: expectedPropertySources,
			}

			req := ht.GET(fmt.Sprintf("/v1/spring/%s/%s/%s", appVersion, appName, devStrProfiles))
			if security {
				applySecurity(t, privKey, req, repo, app)
			}
			res := req.Expect()
			res.JSON().Equal(expectedResult)
		})

		t.Run(fmt.Sprintf("getSpringFileByAppAndProfile_MultiProfile_security:%v", security), func(t *testing.T) {
			flattenApp1IntContent, err := flatten.Flatten(app1IntContent, "", flatten.DotStyle)
			assert.NoError(t, err)
			flattenApp1DevContent, err := flatten.Flatten(app1DevContent, "", flatten.DotStyle)
			assert.NoError(t, err)
			flattenApp1CommonContent, err := flatten.Flatten(app1CommonContent, "", flatten.DotStyle)
			assert.NoError(t, err)
			expectedPropertySources := []*propertySources{
				{
					Name:    "app1-int.yml",
					Source:  flattenApp1IntContent,
					version: commitVersion,
				},
				{
					Name:    "application-int.yml",
					Source:  flattenApp1CommonContent,
					version: commitVersion,
				},
				{
					Name:    "app1-dev.yml",
					Source:  flattenApp1DevContent,
					version: commitVersion,
				},
				{
					Name:    "application-dev.yml",
					Source:  flattenApp1CommonContent,
					version: commitVersion,
				},
				{
					Name:    "app1.yml",
					Source:  flattenApp1CommonContent,
					version: commitVersion,
				},
				{
					Name:    "application.yml",
					Source:  flattenApp1CommonContent,
					version: commitVersion,
				},
			}
			expectedResult := springSummaryResponse{
				Name:            appName,
				Profiles:        devAndIntProfiles,
				Label:           nil,
				Version:         commitVersion,
				State:           nil,
				PropertySources: expectedPropertySources,
			}

			req := ht.GET(fmt.Sprintf("/v1/spring/%s/%s/%s", appVersion, appName, devAndIntStrProfiles))
			if security {
				applySecurity(t, privKey, req, repo, app)
			}
			res := req.Expect()
			res.JSON().Equal(expectedResult)
		})

		mergedConfig := map[string]interface{}{"prop1": app1DevContentPop1, "commonProp": app1CommonContentCommonProp, "common": map[string]interface{}{"subProp": app1DevContentCommonSubProp, "subProp2": app1CommonContentSubProp2}}
		t.Run(fmt.Sprintf("getSpringConfigByAppAndProfile_JSON_security:%v", security), func(t *testing.T) {
			req := ht.GET(fmt.Sprintf("/v1/spring/%s/%s-%s.json", appVersion, appName, devStrProfiles))
			if security {
				applySecurity(t, privKey, req, repo, app)
			}
			res := req.Expect()
			res.JSON().Equal(mergedConfig)
		})
		t.Run(fmt.Sprintf("getSpringConfigByAppAndProfile_YAML_security:%v", security), func(t *testing.T) {
			req := ht.GET(fmt.Sprintf("/v1/spring/%s/%s-%s.yml", appVersion, appName, devStrProfiles))
			if security {
				applySecurity(t, privKey, req, repo, app)
			}
			res := req.Expect()
			rawYml := res.Body().Raw()
			receivedYml := make(map[string]interface{})
			assert.NoError(t, yaml.Unmarshal([]byte(rawYml), &receivedYml))
			assert.Equal(t, mergedConfig, receivedYml)
		})

		if security {
			t.Run("unauthorized", func(t *testing.T) {
				req := ht.GET(fmt.Sprintf("/v1/spring/%s/%s-%s.yml", appVersion, appName, devStrProfiles))
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

func TestServer_Get_InvalidExtension(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepo(ctrl)
	srv := New(repo, "127.0.0.1:8080", false)
	ht := httptest.New(t, srv.app)
	ht.GET("/v1/spring/v1.0.0/app-dev.txt").Expect().Status(httptest.StatusBadRequest)
}

func TestServer_Get_InvalidAppVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepo(ctrl)
	srv := New(repo, "127.0.0.1:8080", false)
	ht := httptest.New(t, srv.app)
	ht.GET("/v1/spring/asdfasd/app-dev.txt").Expect().Status(httptest.StatusBadRequest)
}
