// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package api

import (
	"github.com/ergoapi/exgin"
	"github.com/gin-gonic/gin"
	"next-terminal/constants"
)

func RVersion(c *gin.Context) {
	exgin.GinsData(c, map[string]string{
		"builddate": constants.Date,
		"release":   constants.Release,
		"gitcommit": constants.Commit,
		"version":   constants.Version,
	}, nil)
}
