package api

import (
	"github.com/ergoapi/errors"
	"github.com/ergoapi/exgin"
	"github.com/ergoapi/zlog"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"

	"next-terminal/pkg/global"
	"next-terminal/server/model"
	"next-terminal/server/utils"

)

func UserCreateEndpoint(c *gin.Context) {
	var item model.User
	exgin.Bind(c, &item)
	password := item.Password

	var pass []byte
	var err error
	if pass, err = utils.Encoder.Encode([]byte(password)); err != nil {
		errors.Dangerous(err)
		return
	}
	item.Password = string(pass)

	item.ID = utils.UUID()
	item.Created = utils.NowJsonTime()

	if err := userRepository.Create(&item); err != nil {
		errors.Dangerous(err)
		return
	}

	if item.Mail != "" {
		go mailService.SendMail(item.Mail, "[Next Terminal] 注册通知", "你好，"+item.Nickname+"。管理员为你注册了账号："+item.Username+" 密码："+password)
	}
	Success(c, item)
}

func UserPagingEndpoint(c *gin.Context) {
	pageIndex, _ := strconv.Atoi(c.Query("pageIndex"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	username := c.Query("username")
	nickname := c.Query("nickname")
	mail := c.Query("mail")

	order := c.Query("order")
	field := c.Query("field")

	account, _ := GetCurrentAccount(c)
	items, total, err := userRepository.Find(pageIndex, pageSize, username, nickname, mail, order, field, account)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, H{
		"total": total,
		"items": items,
	})
}

func UserUpdateEndpoint(c *gin.Context) {
	id := c.Param("id")

	var item model.User
	exgin.Bind(c, &item)
	item.ID = id

	if err := userRepository.Update(&item); err != nil {
		errors.Dangerous(err)
		return
	}

	Success(c, nil)
}

func UserDeleteEndpoint(c *gin.Context) {
	ids := c.Param("id")
	account, found := GetCurrentAccount(c)
	if !found {
		Fail(c, -1, "获取当前登录账户失败")
		return
	}
	split := strings.Split(ids, ",")
	for i := range split {
		userId := split[i]
		if account.ID == userId {
			Fail(c, -1, "不允许删除自身账户")
			return
		}
		// 将用户强制下线
		loginLogs, err := loginLogRepository.FindAliveLoginLogsByUserId(userId)
		if err != nil {
			errors.Dangerous(err)
			return
		}

		for j := range loginLogs {
			global.Cache.Delete(loginLogs[j].ID)
			if err := userService.Logout(loginLogs[j].ID); err != nil {
				zlog.Error("%v Cache Deleted Error: %v", loginLogs[j].ID, err)
				Fail(c, 500, "强制下线错误")
				return
			}
		}

		// 删除用户
		if err := userRepository.DeleteById(userId); err != nil {
			errors.Dangerous(err)
			return
		}
	}

	Success(c, nil)
}

func UserGetEndpoint(c *gin.Context) {
	id := c.Param("id")

	item, err := userRepository.FindById(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, item)
}

func UserChangePasswordEndpoint(c *gin.Context) {
	id := c.Param("id")
	password := c.Query("password")

	user, err := userRepository.FindById(id)
	if err != nil {
		errors.Dangerous(err)
		return
	}

	passwd, err := utils.Encoder.Encode([]byte(password))
	if err != nil {
		errors.Dangerous(err)
		return
	}
	u := &model.User{
		Password: string(passwd),
		ID:       id,
	}
	if err := userRepository.Update(u); err != nil {
		errors.Dangerous(err)
		return
	}

	if user.Mail != "" {
		go mailService.SendMail(user.Mail, "[Next Terminal] 密码修改通知", "你好，"+user.Nickname+"。管理员已将你的密码修改为："+password)
	}

	Success(c, "")
}

func UserResetTotpEndpoint(c *gin.Context) {
	id := c.Param("id")
	u := &model.User{
		TOTPSecret: "-",
		ID:         id,
	}
	if err := userRepository.Update(u); err != nil {
		errors.Dangerous(err)
		return
	}
	Success(c, "")
}

func ReloadToken() error {
	loginLogs, err := loginLogRepository.FindAliveLoginLogs()
	if err != nil {
		return err
	}

	for i := range loginLogs {
		loginLog := loginLogs[i]
		token := loginLog.ID
		user, err := userRepository.FindById(loginLog.UserId)
		if err != nil {
			zlog.Debug("用户「%v」获取失败，忽略", loginLog.UserId)
			continue
		}

		authorization := Authorization{
			Token:    token,
			Remember: loginLog.Remember,
			User:     user,
		}

		cacheKey := BuildCacheKeyByToken(token)

		if authorization.Remember {
			// 记住登录有效期两周
			global.Cache.Set(cacheKey, authorization, RememberEffectiveTime)
		} else {
			global.Cache.Set(cacheKey, authorization, NotRememberEffectiveTime)
		}
		zlog.Debug("重新加载用户「%v」授权Token「%v」到缓存", user.Nickname, token)
	}
	return nil
}
