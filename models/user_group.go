package models

type UserGroup struct {
	Model
	ID   string `gorm:"primary_key" json:"id"`
	Name string `json:"name"`
}

type UserGroupForPage struct {
	Model
	ID         string `json:"id"`
	Name       string `json:"name"`
	AssetCount int64  `json:"assetCount"`
}

func (r *UserGroup) TableName() string {
	return "user_groups"
}

type UserGroupMember struct {
	Model
	ID          string `gorm:"primary_key" json:"name"`
	UserId      string `gorm:"index" json:"userId"`
	UserGroupId string `gorm:"index" json:"userGroupId"`
}

func (r *UserGroupMember) TableName() string {
	return "user_group_members"
}

func init() {
	Migrate(UserGroup{})
	Migrate(UserGroupMember{})
}
