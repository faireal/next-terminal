package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"

	"next-terminal/pkg/global"
	"next-terminal/server/model"
	"next-terminal/server/utils"
)

func SecurityCreateEndpoint(c *gin.Context) {
	var item model.AccessSecurity
	exgin.Bind(c, &item)

	item.ID = utils.UUID()
	item.Source = "管理员添加"

	if err := accessSecurityRepository.Create(&item); err != nil {
		errors.Dangerous(err)
		return
	}
	// 更新内存中的安全规则
	if err := ReloadAccessSecurity(); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, "")
}

func ReloadAccessSecurity() error {
	rules, err := accessSecurityRepository.FindAllAccessSecurities()
	if err != nil {
		return err
	}
	if len(rules) > 0 {
		var securities []*global.Security
		for i := 0; i < len(rules); i++ {
			rule := global.Security{
				IP:   rules[i].IP,
				Rule: rules[i].Rule,
			}
			securities = append(securities, &rule)
		}
		global.Securities = securities
	}
	return nil
}

func SecurityPagingEndpoint(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("pageIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	ip := c.Query("ip")
	rule := c.Query("rule")

	order := c.Query("order")
	field := c.Query("field")

	items, total, err := accessSecurityRepository.Find(pageIndex, pageSize, ip, rule, order, field)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, H{
		"total": total,
		"items": items,
	})
}

func SecurityUpdateEndpoint(c *gin.Context) {
	id := c.Param("id")

	var item model.AccessSecurity
	exgin.Bind(c, &item)

	if err := accessSecurityRepository.UpdateById(&item, id); err != nil {
		errors.Dangerous(err)
		return
	}
	// 更新内存中的安全规则
	if err := ReloadAccessSecurity(); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, nil)
}

func SecurityDeleteEndpoint(c *gin.Context) {
	ids := c.Param("id")

	split := strings.Split(ids, ",")
	for i := range split {
		jobId := split[i]
		if err := accessSecurityRepository.DeleteById(jobId); err != nil {
			errors.Dangerous(err)
			return
		}
	}
	// 更新内存中的安全规则
	if err := ReloadAccessSecurity(); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, nil)
}

func SecurityGetEndpoint(c *gin.Context) {
	id := c.Param("id")

	item, err := accessSecurityRepository.FindById(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, item)
}
