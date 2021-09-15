package api

import (
	"fmt"
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"

	"next-terminal/server/model"

	"gorm.io/gorm"
)

func PropertyGetEndpoint(c *gin.Context) {
	properties := propertyRepository.FindAllMap()
	Success(c, properties)
}

func PropertyUpdateEndpoint(c *gin.Context) {
	var item map[string]interface{}
	exgin.Bind(c, &item)

	for key := range item {
		value := fmt.Sprintf("%v", item[key])
		if value == "" {
			value = "-"
		}

		property := model.Configs{
			Ckey: key,
			Cval: value,
		}

		_, err := propertyRepository.FindByName(key)
		if err != nil && err == gorm.ErrRecordNotFound {
			if err := propertyRepository.Create(&property); err != nil {
				errors.Dangerous(err)
				return
			}
		} else {
			if err := propertyRepository.UpdateByName(&property, key); err != nil {
				errors.Dangerous(err)
				return
			}
		}
	}
	Success(c, nil)
}
