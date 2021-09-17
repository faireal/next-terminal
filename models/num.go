package models

type Num struct {
	Model
	I string `gorm:"primary_key" json:"i"`
}

func (r *Num) TableName() string {
	return "nums"
}

func init() {
	Migrate(Num{})
}
