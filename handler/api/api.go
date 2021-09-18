package api

import (
	"github.com/gin-gonic/gin"
	"next-terminal/constants"
	"next-terminal/models"
	ntcache "next-terminal/pkg/cache"
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
	// 1. Authorization JWT Token 临时有效token
	// 2. Token 永久token
	token := c.Request.Header.Get(constants.Token)
	if len(token) > 0 {
		return token
	}
	return c.Query(constants.Token)
}

func GetCurrentAccount(c *gin.Context) (models.User, bool) {
	token := GetToken(c)
	cacheKey := BuildCacheKeyByToken(token)
	get, b := ntcache.MemCache.Get(cacheKey)
	if b {
		return get.(Authorization).User, true
	}
	return models.User{}, false
}

func HasPermission(c *gin.Context, owner string) bool {
	// 检测是否登录
	account, found := GetCurrentAccount(c)
	if !found {
		return false
	}
	// 检测是否为管理人员
	if constants.RoleAdmin == account.Role {
		return true
	}
	// 检测是否为所有者
	if owner == account.ID {
		return true
	}
	return false
}
