package models

// Configs 配置
type Configs struct {
	Model
	Ckey string `json:"ckey"`
	Cval string `json:"cval"`
}

func (c *Configs) TableName() string {
	return "configs"
}

func init() {
	Migrate(Configs{})
}
