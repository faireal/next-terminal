package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	ntcache "next-terminal/pkg/cache"
	"strconv"
	"strings"
)

func LoginLogPagingEndpoint(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("pageIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	userId := c.Query("userId")
	clientIp := c.Query("clientIp")

	items, total, err := loginLogRepository.Find(pageIndex, pageSize, userId, clientIp)

	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, H{
		"total": total,
		"items": items,
	})
}

func LoginLogDeleteEndpoint(c *gin.Context) {
	ids := c.Param("id")
	split := strings.Split(ids, ",")
	for i := range split {
		token := split[i]
		ntcache.MemCache.Delete(token)
		if err := userService.Logout(token); err != nil {
			zlog.Error("Cache Delete Failed")
		}
	}
	if err := loginLogRepository.DeleteByIdIn(split); err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, nil)
}
