// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package models

type Cluster struct {
	Model
	ID          string `gorm:"primary_key " json:"id"`
	Name        string `json:"name"`
	Authtype    string `json:"authtype"`
	Mode        string `json:"mode"`
	Description string `json:"description"`
	Status      bool   `json:"status"`
	Tags        string `json:"tags"`
	Provider    string `json:"provider"`
	Region      string `json:"region"`
	Owner       string `json:"owner"`
	OwnerName   string `json:"ownerName"`
	Kubeconfig  string `json:"kubeconfig"`
	Encrypted   bool   `json:"encrypted"`
}

func (c *Cluster) TableName() string {
	return "clusters"
}

func init() {
	Migrate(Cluster{})
}
