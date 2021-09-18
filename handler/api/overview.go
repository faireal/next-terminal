package api

import (
	"github.com/ergoapi/errors"
	"github.com/gin-gonic/gin"
	"next-terminal/constants"
	"next-terminal/repository"
)

type Counter struct {
	User          int64 `json:"user"`
	Asset         int64 `json:"asset"`
	Credential    int64 `json:"credential"`
	OnlineSession int64 `json:"onlineSession"`
}

func OverviewCounterEndPoint(c *gin.Context) {
	account, _ := GetCurrentAccount(c)

	var (
		countUser          int64
		countOnlineSession int64
		credential         int64
		asset              int64
	)
	if constants.RoleDefault == account.Role {
		countUser, _ = userRepository.CountOnlineUser()
		countOnlineSession, _ = sessionRepository.CountOnlineSession()
		credential, _ = credentialRepository.CountByUserId(account.ID)
		asset, _ = assetRepository.CountByUserId(account.ID)
	} else {
		countUser, _ = userRepository.CountOnlineUser()
		countOnlineSession, _ = sessionRepository.CountOnlineSession()
		credential, _ = credentialRepository.Count()
		asset, _ = assetRepository.Count()
	}
	counter := Counter{
		User:          countUser,
		OnlineSession: countOnlineSession,
		Credential:    credential,
		Asset:         asset,
	}

	Success(c, counter)
}

func OverviewSessionPoint(c *gin.Context) {
	d := c.Query("d")
	var results []repository.D
	var err error
	if d == "m" {
		results, err = sessionRepository.CountSessionByDay(30)
	} else {
		results, err = sessionRepository.CountSessionByDay(7)
	}
	if err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, results)
}
