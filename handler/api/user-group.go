package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"next-terminal/models"
	"next-terminal/pkg/utils"
	"strconv"
	"strings"
)

type UserGroup struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

func UserGroupCreateEndpoint(c *gin.Context) {
	var item UserGroup
	exgin.Bind(c, &item)

	userGroup := models.UserGroup{
		ID:      utils.UUID(),
		Name:    item.Name,
	}

	if err := userGroupRepository.Create(&userGroup, item.Members); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, item)
}

func UserGroupPagingEndpoint(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("pageIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	name := c.Query("name")

	order := c.Query("order")
	field := c.Query("field")

	items, total, err := userGroupRepository.Find(pageIndex, pageSize, name, order, field)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, H{
		"total": total,
		"items": items,
	})
}

func UserGroupUpdateEndpoint(c *gin.Context) {
	id := c.Param("id")

	var item UserGroup
	exgin.Bind(c, &item)
	userGroup := models.UserGroup{
		Name: item.Name,
	}

	if err := userGroupRepository.Update(&userGroup, item.Members, id); err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, nil)
}

func UserGroupDeleteEndpoint(c *gin.Context) {
	ids := c.Param("id")
	split := strings.Split(ids, ",")
	for i := range split {
		userId := split[i]
		if err := userGroupRepository.DeleteById(userId); err != nil {
			errors.Dangerous(err)
			return
		}
	}

	Success(c, nil)
}

func UserGroupGetEndpoint(c *gin.Context) {
	id := c.Param("id")

	item, err := userGroupRepository.FindById(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	members, err := userGroupRepository.FindMembersById(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	userGroup := UserGroup{
		Id:      item.ID,
		Name:    item.Name,
		Members: members,
	}

	Success(c, userGroup)
}
