// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package api

import (
	"encoding/json"
	"fmt"
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/ergoapi/util/zos"
	"github.com/gin-gonic/gin"
	"next-terminal/models"
	"next-terminal/pkg/utils"
	"strconv"
	"strings"
)

// ClusterGetAll 获取集群列表
func ClusterGetAll(c *gin.Context) {
	account, _ := GetCurrentAccount(c)
	items, _ := clusterRepository.FindByProtocolAndUser(account)
	exgin.GinsData(c, items, nil)
}

// ClusterCreate 新增集群
func ClusterCreate(c *gin.Context) {
	m := map[string]interface{}{}
	exgin.Bind(c, &m)

	data, _ := json.Marshal(m)
	var item models.Cluster
	if err := json.Unmarshal(data, &item); err != nil {
		errors.Dangerous(err)
		return
	}

	account, _ := GetCurrentAccount(c)
	item.Owner = account.ID
	item.ID = zos.GenUUID()

	if err := clusterRepository.InitCluster(&item, m); err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, item, nil)
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

// ClusterGet cluster
func ClusterGet(c *gin.Context) {
	id := c.Param("id")
	if err := PreCheckClusterPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	var item *models.Cluster
	var err error
	if item, err = clusterRepository.Get("id = ?", id); err != nil {
		errors.Dangerous(err)
		return
	}
	itemMap := utils.StructToMap(item)
	exgin.GinsData(c, itemMap, nil)
}

func ClusterUpdate(c *gin.Context) {
	id := c.Param("id")
	if err := PreCheckClusterPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	m := map[string]interface{}{}
	exgin.Bind(c, &m)

	data, _ := json.Marshal(m)
	var item models.Cluster
	if err := json.Unmarshal(data, &item); err != nil {
		errors.Dangerous(err)
		return
	}

	if len(item.Tags) == 0 {
		item.Tags = "-"
	}

	if item.Description == "" {
		item.Description = "-"
	}

	if err := clusterRepository.Encrypt(&item, utils.Encryption()); err != nil {
		errors.Dangerous(err)
		return
	}
	if err := clusterRepository.Update(&item, id); err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, nil, nil)
}

func PreCheckClusterPermission(c *gin.Context, id string) error {
	item, err := clusterRepository.Get("id = ?", id)
	if err != nil {
		return err
	}

	if !HasPermission(c, item.Owner) {
		return fmt.Errorf("permission denied")
	}
	return nil
}

func ClusterPingEndpoint(c *gin.Context) {
	id := c.Param("id")

	var item *models.Cluster
	var err error
	if item, err = clusterRepository.Get("id = ?", id); err != nil {
		errors.Dangerous(err)
		return
	}

	active := utils.KubePing(item.Kubeconfig, item.ID)

	if item.Status != active {
		if err := clusterRepository.UpdateStatusByID(active, item.ID); err != nil {
			errors.Dangerous(err)
			return
		}
	}

	exgin.GinsData(c, active, nil)
}

func ClusterDelete(c *gin.Context) {
	id := c.Param("id")
	split := strings.Split(id, ",")
	for i := range split {
		if err := PreCheckClusterPermission(c, split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
		if err := clusterRepository.DeleteByID(split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
		// 删除资产与用户的关系
		// TODO
	}

	exgin.GinsData(c, nil, nil)
}
