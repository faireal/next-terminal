// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"strconv"
)

// ClusterGetAll 获取集群列表
func ClusterGetAll(c *gin.Context) {
	account, _ := GetCurrentAccount(c)
	items, _ := clusterRepository.FindByProtocolAndUser(account)
	exgin.GinsData(c, items, nil)
}

// ClusterPagingEndpoint cluster
func ClusterPagingEndpoint(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("pageIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	name := c.Query("name")
	tags := c.Query("tags")
	owner := c.Query("owner")

	order := c.Query("order")
	field := c.Query("field")

	account, _ := GetCurrentAccount(c)
	items, total, err := clusterRepository.Find(pageIndex, pageSize, name, tags, account, owner, order, field)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, H{
		"total": total,
		"items": items,
	}, nil)
}