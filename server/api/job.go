package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"

	"next-terminal/server/model"
	"next-terminal/server/utils"
)

func JobCreateEndpoint(c *gin.Context) {
	var item model.Job
	exgin.Bind(c, &item)
	item.ID = utils.UUID()
	item.Created = utils.NowJsonTime()

	if err := jobService.Create(&item); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, "")
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

	Success(c, H{
		"total": total,
		"items": items,
	})
}

func JobUpdateEndpoint(c *gin.Context) {
	id := c.Param("id")

	var item model.Job
	exgin.Bind(c, &item)
	item.ID = id
	if err := jobRepository.UpdateById(&item); err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, nil)
}

func JobChangeStatusEndpoint(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status")
	if err := jobService.ChangeStatusById(id, status); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, "")
}

func JobExecEndpoint(c *gin.Context) {
	id := c.Param("id")
	if err := jobService.ExecJobById(id); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, "")
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

	Success(c, nil)
}

func JobGetEndpoint(c *gin.Context) {
	id := c.Param("id")

	item, err := jobRepository.FindById(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, item)
}

func JobGetLogsEndpoint(c *gin.Context) {
	id := c.Param("id")

	items, err := jobLogRepository.FindByJobId(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, items)
}

func JobDeleteLogsEndpoint(c *gin.Context) {
	id := c.Param("id")
	if err := jobLogRepository.DeleteByJobId(id); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, "")
}
