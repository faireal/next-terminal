package api

import (
	"fmt"
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"next-terminal/models"

	"gorm.io/gorm"
)

func PropertyGetEndpoint(c *gin.Context) {
	properties := configsRepository.FindAllMap()
	exgin.GinsData(c, properties, nil)
}

func PropertyUpdateEndpoint(c *gin.Context) {
	var item map[string]interface{}
	exgin.Bind(c, &item)

	for key := range item {
		value := fmt.Sprintf("%v", item[key])
		if value == "" {
			value = "-"
		}

		property := models.Configs{
			Ckey: key,
			Cval: value,
		}

		_, err := configsRepository.FindByName(key)
		if err != nil && err == gorm.ErrRecordNotFound {
			if err := configsRepository.Create(&property); err != nil {
				errors.Dangerous(err)
				return
			}
		} else {
			if err := configsRepository.UpdateByName(&property, key); err != nil {
				errors.Dangerous(err)
				return
			}
		}
	}
	exgin.GinsData(c, nil, nil)
}

func ShowCfg(c *gin.Context) {
	cfg := map[string]interface{}{}
	cfg["cluster"] = true
	exgin.GinsData(c, cfg, nil)
}
