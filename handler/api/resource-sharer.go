package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
)

type RU struct {
	UserGroupId  string   `json:"userGroupId"`
	UserId       string   `json:"userId"`
	ResourceType string   `json:"resourceType"`
	ResourceIds  []string `json:"resourceIds"`
}

type UR struct {
	ResourceId   string   `json:"resourceId"`
	ResourceType string   `json:"resourceType"`
	UserIds      []string `json:"userIds"`
}

func RSGetSharersEndPoint(c *gin.Context) {
	resourceId := c.Query("resourceId")
	userIds, err := resourceSharerRepository.FindUserIdsByResourceId(resourceId)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, userIds, nil)
}

func RSOverwriteSharersEndPoint(c *gin.Context) {
	var ur UR
	exgin.Bind(c, &ur)

	if err := resourceSharerRepository.OverwriteUserIdsByResourceId(ur.ResourceId, ur.ResourceType, ur.UserIds); err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, "", nil)
}

func ResourceRemoveByUserIdAssignEndPoint(c *gin.Context) {
	var ru RU
	exgin.Bind(c, &ru)

	if err := resourceSharerRepository.DeleteByUserIdAndResourceTypeAndResourceIdIn(ru.UserGroupId, ru.UserId, ru.ResourceType, ru.ResourceIds); err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, "", nil)
}

func ResourceAddByUserIdAssignEndPoint(c *gin.Context) {
	var ru RU
	exgin.Bind(c, &ru)

	if err := resourceSharerRepository.AddSharerResources(ru.UserGroupId, ru.UserId, ru.ResourceType, ru.ResourceIds); err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, "", nil)
}
