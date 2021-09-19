package models

import (
	"gopkg.in/guregu/null.v3"
)

type LoginLog struct {
	Model
	ID              string         `gorm:"primary_key" json:"id"`
	UserId          string         `gorm:"index" json:"userId"`
	ClientIP        string         `json:"clientIp"`
	ClientUserAgent string         `json:"clientUserAgent"`
	LoginTime       null.Time `json:"loginTime"`
	LogoutTime      null.Time `json:"logoutTime"`
	Remember        bool           `json:"remember"`
}

type LoginLogForPage struct {
	Model
	ID              string         `json:"id"`
	UserId          string         `json:"userId"`
	UserName        string         `json:"userName"`
	ClientIP        string         `json:"clientIp"`
	ClientUserAgent string         `json:"clientUserAgent"`
	LoginTime       null.Time `json:"loginTime"`
	LogoutTime      null.Time `json:"logoutTime"`
	Remember        bool           `json:"remember"`
}

func (r *LoginLog) TableName() string {
	return "login_logs"
}

func init() {
	Migrate(LoginLog{})
}
