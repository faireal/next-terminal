package models

import (
	"gopkg.in/guregu/null.v3"
)

type Logs struct {
	Model
	ID              string    `gorm:"primary_key" json:"id"`
	UserId          string    `gorm:"index" json:"userId"`
	ClientIP        string    `json:"clientIp"`
	ClientUserAgent string    `json:"clientUserAgent"`
	LoginTime       null.Time `json:"loginTime"`
	LogoutTime      null.Time `json:"logoutTime"`
	ActionTime      null.Time `json:"ActionTime"`
	Remember        bool      `json:"remember"`
	Action          string    `json:"action"`
}

type LogsForPage struct {
	Model
	ID              string    `json:"id"`
	UserId          string    `json:"userId"`
	UserName        string    `json:"userName"`
	ClientIP        string    `json:"clientIp"`
	ClientUserAgent string    `json:"clientUserAgent"`
	LoginTime       null.Time `json:"loginTime"`
	LogoutTime      null.Time `json:"logoutTime"`
	ActionTime      null.Time `json:"ActionTime"`
	Remember        bool      `json:"remember"`
	Action          string    `json:"action"`
}

func (r *Logs) TableName() string {
	return "logs"
}

func init() {
	Migrate(Logs{})
}
