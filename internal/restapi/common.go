package restapi

import (
	"github.com/kataras/iris"
	"github.com/sirupsen/logrus"
	"github.com/vecosy/vecosy/v2/internal/utils"
)

const invalidFormatErrorMessage = "unsupported extension. Valid formats: [.yml,.json]"

func respondConfig(ctx iris.Context, finalConfig map[interface{}]interface{}, ext string, log *logrus.Entry) {
	// converting and responding
	var err error
	switch ext {
	case ".yml", ".yaml":
		_, err = ctx.YAML(finalConfig)
	case ".json":
		normalizedMap, err := utils.NormalizeMap(finalConfig)
		if err != nil {
			log.Errorf("Error normalizing json map:%#+vs, err:%s", finalConfig, err)
			internalServerError(ctx)
			return
		}
		_, err = ctx.JSON(normalizedMap)
	default:
		badRequest(ctx, invalidFormatErrorMessage)
		return
	}
	if err != nil {
		log.Errorf("Error responding :%s", err)
		internalServerError(ctx)
	}
}
