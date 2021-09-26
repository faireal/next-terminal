// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package models

type Cluster struct {
	Model
	ID          string `gorm:"primary_key " json:"id"`
	Name        string `json:"name"`
	Mode        string `json:"mode"`
	Description string `json:"description,omitempty"`
	Status      bool   `json:"status"`
	Tags        string `json:"tags,omitempty"`
	Provider    string `json:"provider,omitempty"`
	Region      string `json:"region,omitempty"`
	Owner       string `json:"owner"`
	OwnerName   string `json:"ownerName"`
}

func (c *Cluster) TableName() string {
	return "clusters"
}

func init() {
	Migrate(Cluster{})
}
