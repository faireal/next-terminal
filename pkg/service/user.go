package service

import (
	"github.com/ergoapi/zlog"
	"next-terminal/constants"
	"next-terminal/models"
	"next-terminal/pkg/utils"
	"next-terminal/repository"
)

type UserService struct {
	userRepository *repository.UserRepository
	logsRepository *repository.LogsRepository
}

func NewUserService(userRepository *repository.UserRepository, logsRepository *repository.LogsRepository) *UserService {
	return &UserService{userRepository: userRepository, logsRepository: logsRepository}
}

func (r UserService) InitUser() (err error) {

	users := r.userRepository.FindAll()

	if len(users) == 0 {
		initPassword := utils.GetKeyFromYaml("core.admin.password", "admin")
		var pass []byte
		if pass, err = utils.Encoder.Encode([]byte(initPassword)); err != nil {
			return err
		}

		user := models.User{
			ID:       utils.UUID(),
			Username: utils.GetKeyFromYaml("core.admin.username", "admin"),
			Password: string(pass),
			Nickname: "超级管理员",
			Role:     constants.RoleAdmin,
		}
		if err := r.userRepository.Create(&user); err != nil {
			return err
		}
		zlog.Info("初始用户创建成功，账号：「%v」密码：「%v」", user.Username, initPassword)
	} else {
		for i := range users {
			// 修正默认用户类型为普通用户
			if users[i].Role == "" {
				user := models.User{
					Role: constants.RoleDefault,
					ID:   users[i].ID,
				}
				if err := r.userRepository.Update(user); err != nil {
					return err
				}
				zlog.Info("自动修正用户「%v」ID「%v」类型为普通用户", users[i].Nickname, users[i].ID)
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
			logs, err := r.logsRepository.FindAliveLoginLogsByUserId(onlineUsers[i].ID)
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

	loginLog, err := r.logsRepository.FindById(token)
	if err != nil {
		zlog.Warn("登录日志「%v」获取失败", token)
		return
	}

	loginLogForUpdate := &models.Logs{ID: token}
	err = r.logsRepository.Update(loginLogForUpdate)
	if err != nil {
		return err
	}

	loginLogs, err := r.logsRepository.FindAliveLoginLogsByUserId(loginLog.UserId)
	if err != nil {
		return
	}

	if len(loginLogs) == 0 {
		err = r.userRepository.UpdateOnline(loginLog.UserId, false)
	}
	return
}
