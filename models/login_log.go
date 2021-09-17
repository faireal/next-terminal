package models

import (
	"next-terminal/pkg/utils"
)

type LoginLog struct {
	Model
	ID              string         `gorm:"primary_key" json:"id"`
	UserId          string         `gorm:"index" json:"userId"`
	ClientIP        string         `json:"clientIp"`
	ClientUserAgent string         `json:"clientUserAgent"`
	LoginTime       utils.JsonTime `json:"loginTime"`
	LogoutTime      utils.JsonTime `json:"logoutTime"`
	Remember        bool           `json:"remember"`
}

type LoginLogForPage struct {
	ID              string         `json:"id"`
	UserId          string         `json:"userId"`
	UserName        string         `json:"userName"`
	ClientIP        string         `json:"clientIp"`
	ClientUserAgent string         `json:"clientUserAgent"`
	LoginTime       utils.JsonTime `json:"loginTime"`
	LogoutTime      utils.JsonTime `json:"logoutTime"`
	Remember        bool           `json:"remember"`
}

func (r *LoginLog) TableName() string {
	return "login_logs"
}

func init() {
	Migrate(LoginLog{})
}
