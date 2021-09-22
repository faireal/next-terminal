package models

type Asset struct {
	Model
	ID           string `gorm:"primary_key " json:"id"`
	Name         string `json:"name"`
	Protocol     string `json:"protocol"`
	IP           string `json:"ip"`
	Port         int    `json:"port"`
	AccountType  string `json:"accountType"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CredentialId string `gorm:"index" json:"credentialId"`
	PrivateKey   string `json:"privateKey"`
	Passphrase   string `json:"passphrase"`
	Description  string `json:"description"`
	Active       bool   `json:"active"`
	Tags         string `json:"tags"`
	Owner        string `gorm:"index" json:"owner"`
	Encrypted    bool   `json:"encrypted"`
	Provider     string `json:"provider"`
}

type AssetForPage struct {
	Model
	ID          string `json:"id"`
	Name        string `json:"name"`
	IP          string `json:"ip"`
	Protocol    string `json:"protocol"`
	Port        int    `json:"port"`
	Active      bool   `json:"active"`
	Tags        string `json:"tags"`
	Owner       string `json:"owner"`
	OwnerName   string `json:"ownerName"`
	SharerCount int64  `json:"sharerCount"`
	SshMode     string `json:"sshMode"`
	Provider    string `json:"provider"`
}

func (r *Asset) TableName() string {
	return "assets"
}

type AssetAttribute struct {
	Model
	Id      string `gorm:"index" json:"id"`
	AssetId string `gorm:"index" json:"assetId"`
	Name    string `gorm:"index" json:"name"`
	Value   string `json:"value"`
}

func (r *AssetAttribute) TableName() string {
	return "asset_attributes"
}

func init() {
	Migrate(Asset{})
	Migrate(AssetAttribute{})
}
