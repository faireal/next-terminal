package model

// Configs 配置
type Configs struct {
	Ckey string `json:"ckey"`
	Cval string `json:"cval"`
}

func (c *Configs) TableName() string {
	return "configs"
}
