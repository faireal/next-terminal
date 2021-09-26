package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/ergoapi/util/zos"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"next-terminal/models"
	"strconv"
	"strings"

	"next-terminal/pkg/global"
)

func SecurityCreateEndpoint(c *gin.Context) {
	var item models.AccessSecurity
	exgin.Bind(c, &item)

	item.ID = zos.GenUUID()
	item.Source = "管理员添加"

	if viper.GetBool("demo") {
		if strings.Contains(item.IP, "/") && !strings.Contains(item.IP, "/32") {
			errors.Dangerous("演示模式下, 仅允许操作单ip")
			return
		}
	}

	if err := accessSecurityRepository.Create(&item); err != nil {
		errors.Dangerous(err)
		return
	}
	// 更新内存中的安全规则
	if err := ReloadAccessSecurity(); err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, "", nil)
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

	exgin.GinsData(c, H{
		"total": total,
		"items": items,
	}, nil)
}

func SecurityUpdateEndpoint(c *gin.Context) {
	id := c.Param("id")

	var item models.AccessSecurity
	exgin.Bind(c, &item)

	if viper.GetBool("demo") {
		if strings.Contains(item.IP, "/") && !strings.Contains(item.IP, "/32") {
			errors.Dangerous("演示模式下, 仅允许操作单ip")
			return
		}
	}

	if err := accessSecurityRepository.UpdateByID(&item, id); err != nil {
		errors.Dangerous(err)
		return
	}
	// 更新内存中的安全规则
	if err := ReloadAccessSecurity(); err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, nil, nil)
}

func SecurityDeleteEndpoint(c *gin.Context) {
	ids := c.Param("id")

	split := strings.Split(ids, ",")
	for i := range split {
		jobId := split[i]
		if err := accessSecurityRepository.DeleteByID(jobId); err != nil {
			errors.Dangerous(err)
			return
		}
	}
	// 更新内存中的安全规则
	if err := ReloadAccessSecurity(); err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, nil, nil)
}

func SecurityGetEndpoint(c *gin.Context) {
	id := c.Param("id")

	item, err := accessSecurityRepository.FindByID(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, item, nil)
}
