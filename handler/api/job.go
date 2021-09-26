package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/ergoapi/util/zos"
	"github.com/gin-gonic/gin"
	"next-terminal/models"
	"strconv"
	"strings"
)

func JobCreateEndpoint(c *gin.Context) {
	var item models.Job
	exgin.Bind(c, &item)
	item.ID = zos.GenUUID()

	if err := jobService.Create(&item); err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, "", nil)
}

func JobPagingEndpoint(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("pageIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	name := c.Query("name")
	status := c.Query("status")

	order := c.Query("order")
	field := c.Query("field")

	items, total, err := jobRepository.Find(pageIndex, pageSize, name, status, order, field)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, H{
		"total": total,
		"items": items,
	}, nil)
}

func JobUpdateEndpoint(c *gin.Context) {
	id := c.Param("id")

	var item models.Job
	exgin.Bind(c, &item)
	item.ID = id
	if err := jobRepository.UpdateById(&item); err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, nil, nil)
}

func JobChangeStatusEndpoint(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status")
	if err := jobService.ChangeStatusById(id, status); err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, "", nil)
}

func JobExecEndpoint(c *gin.Context) {
	id := c.Param("id")
	if err := jobService.ExecJobById(id); err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, "", nil)
}

func JobDeleteEndpoint(c *gin.Context) {
	ids := c.Param("id")

	split := strings.Split(ids, ",")
	for i := range split {
		jobId := split[i]
		if err := jobRepository.DeleteJobById(jobId); err != nil {
			errors.Dangerous(err)
			return
		}
	}

	exgin.GinsData(c, nil, nil)
}

func JobGetEndpoint(c *gin.Context) {
	id := c.Param("id")

	item, err := jobRepository.FindById(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, item, nil)
}

func JobGetLogsEndpoint(c *gin.Context) {
	id := c.Param("id")

	items, err := jobLogRepository.FindByJobId(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	exgin.GinsData(c, items, nil)
}

func JobDeleteLogsEndpoint(c *gin.Context) {
	id := c.Param("id")
	if err := jobLogRepository.DeleteByJobId(id); err != nil {
		errors.Dangerous(err)
		return
	}
	exgin.GinsData(c, "", nil)
}
