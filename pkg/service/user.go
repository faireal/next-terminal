package service

import (
	"github.com/ergoapi/zlog"
	model2 "next-terminal/models"
	"next-terminal/pkg/constant"
	"next-terminal/pkg/utils"
	repository2 "next-terminal/repository"
)

type UserService struct {
	userRepository     *repository2.UserRepository
	loginLogRepository *repository2.LoginLogRepository
}

func NewUserService(userRepository *repository2.UserRepository, loginLogRepository *repository2.LoginLogRepository) *UserService {
	return &UserService{userRepository: userRepository, loginLogRepository: loginLogRepository}
}

func (r UserService) InitUser() (err error) {

	users := r.userRepository.FindAll()

	if len(users) == 0 {
		initPassword := "admin"
		var pass []byte
		if pass, err = utils.Encoder.Encode([]byte(initPassword)); err != nil {
			return err
		}

		user := model2.User{
			ID:       utils.UUID(),
			Username: "admin",
			Password: string(pass),
			Nickname: "超级管理员",
			Type:     constant.TypeAdmin,
			Created:  utils.NowJsonTime(),
		}
		if err := r.userRepository.Create(&user); err != nil {
			return err
		}
		zlog.Info("初始用户创建成功，账号：「%v」密码：「%v」", user.Username, initPassword)
	} else {
		for i := range users {
			// 修正默认用户类型为管理员
			if users[i].Type == "" {
				user := model2.User{
					Type: constant.TypeAdmin,
					ID:   users[i].ID,
				}
				if err := r.userRepository.Update(&user); err != nil {
					return err
				}
				zlog.Info("自动修正用户「%v」ID「%v」类型为管理员", users[i].Nickname, users[i].ID)
			}
		}
	}
	return nil
}

func (r UserService) FixUserOnlineState() error {
	// 修正用户登录状态
	onlineUsers, err := r.userRepository.FindOnlineUsers()
	if err != nil {
		return err
	}
	if len(onlineUsers) > 0 {
		for i := range onlineUsers {
			logs, err := r.loginLogRepository.FindAliveLoginLogsByUserId(onlineUsers[i].ID)
			if err != nil {
				return err
			}
			if len(logs) == 0 {
				if err := r.userRepository.UpdateOnline(onlineUsers[i].ID, false); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r UserService) Logout(token string) (err error) {

	loginLog, err := r.loginLogRepository.FindById(token)
	if err != nil {
		zlog.Warn("登录日志「%v」获取失败", token)
		return
	}

	loginLogForUpdate := &model2.LoginLog{LogoutTime: utils.NowJsonTime(), ID: token}
	err = r.loginLogRepository.Update(loginLogForUpdate)
	if err != nil {
		return err
	}

	loginLogs, err := r.loginLogRepository.FindAliveLoginLogsByUserId(loginLog.UserId)
	if err != nil {
		return
	}

	if len(loginLogs) == 0 {
		err = r.userRepository.UpdateOnline(loginLog.UserId, false)
	}
	return
}
