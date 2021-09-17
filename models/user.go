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
	Enabled    bool           `json:"enabled"`
	Created    utils.JsonTime `json:"created"`
	Type       string         `json:"type"`
	Mail       string         `json:"mail"`
}

type UserForPage struct {
	ID               string         `json:"id"`
	Username         string         `json:"username"`
	Nickname         string         `json:"nickname"`
	TOTPSecret       string         `json:"totpSecret"`
	Mail             string         `json:"mail"`
	Online           bool           `json:"online"`
	Enabled          bool           `json:"enabled"`
	Created          utils.JsonTime `json:"created"`
	Type             string         `json:"type"`
	SharerAssetCount int64          `json:"sharerAssetCount"`
}

func (r *User) TableName() string {
	return "users"
}

func init() {
	Migrate(User{})
}
