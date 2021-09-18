package models

import (
	"next-terminal/pkg/utils"
)

type User struct {
	Model
	ID         string         `gorm:"primary_key" json:"id"`
	Username   string         `gorm:"index" json:"username"`
	Password   string         `json:"password"`
	Nickname   string         `json:"nickname"`
	TOTPSecret string         `json:"-"`
	Online     bool           `json:"online"`
	Created    utils.JsonTime `json:"created"`
	Role       string         `json:"role"`
	Mail       string         `json:"mail"`
	Baned      bool           `json:"baned"`
	Department string         `json:"department"`
	Mode       string         `json:"mode"` // ldap,local,github
}

type UserForPage struct {
	Model
	ID               string         `json:"id"`
	Username         string         `json:"username"`
	Nickname         string         `json:"nickname"`
	TOTPSecret       string         `json:"totpSecret"`
	Mail             string         `json:"mail"`
	Online           bool           `json:"online"`
	Created          utils.JsonTime `json:"created"`
	Role             string         `json:"role"`
	Department       string         `json:"department"`
	Mode             string         `json:"mode"` // ldap,local,github
	Baned            bool           `json:"baned"`
	SharerAssetCount int64          `json:"sharerAssetCount"`
}

func (r *User) TableName() string {
	return "users"
}

func init() {
	Migrate(User{})
}
