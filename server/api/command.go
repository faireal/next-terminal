package api

import (
	"fmt"
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"

	"next-terminal/server/model"
	"next-terminal/server/utils"
)

func CommandCreateEndpoint(c *gin.Context) {
	var item model.Command
	exgin.Bind(c, &item)
	account, _ := GetCurrentAccount(c)
	item.Owner = account.ID
	item.ID = utils.UUID()
	item.Created = utils.NowJsonTime()

	if err := commandRepository.Create(&item); err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, item)
}

func CommandPagingEndpoint(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("pageIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	name := c.Query("name")
	content := c.Query("content")
	account, _ := GetCurrentAccount(c)

	order := c.Query("order")
	field := c.Query("field")

	items, total, err := commandRepository.Find(pageIndex, pageSize, name, content, order, field, account)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, H{
		"total": total,
		"items": items,
	})
}

func CommandUpdateEndpoint(c *gin.Context) {
	id := c.Param("id")
	if err := PreCheckCommandPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	var item model.Command
	exgin.Bind(c, &item)

	if err := commandRepository.UpdateById(&item, id); err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, nil)
}

func CommandDeleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	split := strings.Split(id, ",")
	for i := range split {
		if err := PreCheckCommandPermission(c, split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
		if err := commandRepository.DeleteById(split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
		// 删除资产与用户的关系
		if err := resourceSharerRepository.DeleteResourceSharerByResourceId(split[i]); err != nil {
			errors.Dangerous(err)
			return
		}
	}
	Success(c, nil)
}

func CommandGetEndpoint(c *gin.Context) {
	id := c.Param("id")

	if err := PreCheckCommandPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	var item model.Command
	var err error
	if item, err = commandRepository.FindById(id); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, item)
}

func CommandChangeOwnerEndpoint(c *gin.Context) {
	id := c.Param("id")

	if err := PreCheckCommandPermission(c, id); err != nil {
		errors.Dangerous(err)
		return
	}

	owner := c.Query("owner")
	if err := commandRepository.UpdateById(&model.Command{Owner: owner}, id); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, "")
}

func PreCheckCommandPermission(c *gin.Context, id string) error {
	item, err := commandRepository.FindById(id)
	if err != nil {
		return err
	}

	if !HasPermission(c, item.Owner) {
		return fmt.Errorf("permission denied")
	}
	return nil
}
