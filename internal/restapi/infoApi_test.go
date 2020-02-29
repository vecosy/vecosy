package restapi

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-version"
	"github.com/kataras/iris/httptest"
	"github.com/vecosy/vecosy/v2/mocks"
	"testing"
)

func TestRest_Info(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepo(ctrl)
	srv := New(repo, "127.0.0.1:8080", false)
	ht := httptest.New(t, srv.app)

	v100, _ := version.NewVersion("1.0.0")
	v101, _ := version.NewVersion("1.0.1")
	v200, _ := version.NewVersion("2.0.0")
	v300, _ := version.NewVersion("3.0.0")
	apps := map[string][]*version.Version{
		"app1": {v100, v101, v200},
		"app2": {v300},
	}
	expected := map[string][]string{
		"app1": {v100.String(), v101.String(), v200.String()},
		"app2": {v300.String()},
	}

	repo.EXPECT().GetAppsVersions().Return(apps)
	res := ht.GET("/v1/info/").Expect()
	res.JSON().Equal(expected)
}

func TestRest_Info_GetApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepo(ctrl)
	srv := New(repo, "127.0.0.1:8080", false)
	ht := httptest.New(t, srv.app)

	v100, _ := version.NewVersion("1.0.0")
	v101, _ := version.NewVersion("1.0.1")
	v200, _ := version.NewVersion("2.0.0")
	v300, _ := version.NewVersion("3.0.0")
	apps := map[string][]*version.Version{
		"app1": {v100, v101, v200},
		"app2": {v300},
	}
	expected := []string{v100.String(), v101.String(), v200.String()}

	repo.EXPECT().GetAppsVersions().Return(apps)
	res := ht.GET("/v1/info/app1").Expect()
	res.JSON().Equal(expected)
}

func TestRest_Info_GetApp_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepo(ctrl)
	srv := New(repo, "127.0.0.1:8080", false)
	ht := httptest.New(t, srv.app)

	v100, _ := version.NewVersion("1.0.0")
	v101, _ := version.NewVersion("1.0.1")
	v200, _ := version.NewVersion("2.0.0")
	v300, _ := version.NewVersion("3.0.0")
	apps := map[string][]*version.Version{
		"app1": {v100, v101, v200},
		"app2": {v300},
	}
	expected := []string{}
	repo.EXPECT().GetAppsVersions().Return(apps)
	res := ht.GET("/v1/info/not_existentApp").Expect()
	res.JSON().Equal(expected)
}
