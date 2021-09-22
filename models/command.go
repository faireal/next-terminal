package models

type Command struct {
	Model
	ID      string `gorm:"primary_key" json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Owner   string `gorm:"index" json:"owner"`
}

type CommandForPage struct {
	ID          string `gorm:"primary_key" json:"id"`
	Name        string `json:"name"`
	Content     string `json:"content"`
	Owner       string `json:"owner"`
	OwnerName   string `json:"ownerName"`
	SharerCount int64  `json:"sharerCount"`
}

func (r *Command) TableName() string {
	return "commands"
}

func init() {
	Migrate(Command{})
}
