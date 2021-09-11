package api

import (
	"github.com/gin-gonic/gin"
	"next-terminal/pkg/constant"
	"next-terminal/pkg/global"
	"next-terminal/server/model"
)

type H map[string]interface{}

func Fail(c *gin.Context, code int, message string) {
	c.JSON(200, H{
		"code":    code,
		"message": message,
	})
}

func FailWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(200, H{
		"code":    code,
		"message": message,
		"data":    data,
	})
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, H{
		"code":    1,
		"message": "success",
		"data":    data,
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(200, H{
		"code":    -1,
		"message": message,
	})
}

func GetToken(c *gin.Context) string {
	token := c.Request.Header.Get(Token)
	if len(token) > 0 {
		return token
	}
	return c.Query(Token)
}

func GetCurrentAccount(c *gin.Context) (model.User, bool) {
	token := GetToken(c)
	cacheKey := BuildCacheKeyByToken(token)
	get, b := global.Cache.Get(cacheKey)
	if b {
		return get.(Authorization).User, true
	}
	return model.User{}, false
}

func HasPermission(c *gin.Context, owner string) bool {
	// 检测是否登录
	account, found := GetCurrentAccount(c)
	if !found {
		return false
	}
	// 检测是否为管理人员
	if constant.TypeAdmin == account.Type {
		return true
	}
	// 检测是否为所有者
	if owner == account.ID {
		return true
	}
	return false
}
