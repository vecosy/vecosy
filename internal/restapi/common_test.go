package restapi

import (
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/mocks"
	"net/http"
	"testing"
)

func Test_respondConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := logrus.NewEntry(nil)
	ctx := mocks.NewMockContext(ctrl)
	config := map[interface{}]interface{}{
		"config": "test",
	}
	ctx.EXPECT().YAML(gomock.Eq(config)).Times(2)
	respondConfig(ctx, config, ".yaml", log)
	respondConfig(ctx, config, ".yml", log)

	ctx.EXPECT().JSON(gomock.Any()).Times(1)
	respondConfig(ctx, config, ".json", log)

	ctx.EXPECT().WriteString(invalidFormatErrorMessage).Times(1)
	ctx.EXPECT().StatusCode(gomock.Eq(http.StatusBadRequest)).Times(1)
	respondConfig(ctx, config, ".notValid", log)
}
