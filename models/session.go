package models

import (
	"gopkg.in/guregu/null.v3"
)

type Session struct {
	Model
	ID               string    `gorm:"primary_key" json:"id"`
	Protocol         string    `json:"protocol"`
	IP               string    `json:"ip"`
	Port             int       `json:"port"`
	ConnectionId     string    `json:"connectionId"`
	AssetId          string    `gorm:"index" json:"assetId"`
	Username         string    `json:"username"`
	Password         string    `json:"password"`
	Creator          string    `gorm:"index" json:"creator"`
	ClientIP         string    `json:"clientIp"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	Status           string    `gorm:"index" json:"status"`
	Recording        string    `json:"recording"`
	PrivateKey       string    `json:"privateKey"`
	Passphrase       string    `json:"passphrase"`
	Code             int       `json:"code"`
	Message          string    `json:"message"`
	ConnectedTime    null.Time `json:"connectedTime"`
	DisconnectedTime null.Time `json:"disconnectedTime"`
	Mode             string    `json:"mode"`
}

func (r *Session) TableName() string {
	return "sessions"
}

type SessionForPage struct {
	Model
	ID               string    `json:"id"`
	Protocol         string    `json:"protocol"`
	IP               string    `json:"ip"`
	Port             int       `json:"port"`
	Username         string    `json:"username"`
	ConnectionId     string    `json:"connectionId"`
	AssetId          string    `json:"assetId"`
	Creator          string    `json:"creator"`
	ClientIP         string    `json:"clientIp"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	Status           string    `json:"status"`
	Recording        string    `json:"recording"`
	ConnectedTime    null.Time `json:"connectedTime"`
	DisconnectedTime null.Time `json:"disconnectedTime"`
	AssetName        string    `json:"assetName"`
	CreatorName      string    `json:"creatorName"`
	Code             int       `json:"code"`
	Message          string    `json:"message"`
	Mode             string    `json:"mode"`
}

func init() {
	Migrate(Session{})
}
