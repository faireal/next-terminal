package repository

import (
	"encoding/base64"
	"fmt"
	"github.com/ergoapi/zlog"
	"github.com/spf13/viper"
	"next-terminal/constants"
	"next-terminal/models"
	"next-terminal/pkg/utils"
	"strings"

	"gorm.io/gorm"
)

type AssetRepository struct {
	DB *gorm.DB
}

func NewAssetRepository(db *gorm.DB) *AssetRepository {
	assetRepository = &AssetRepository{DB: db}
	return assetRepository
}

func (r AssetRepository) FindAll() (o []models.Asset, err error) {
	err = r.DB.Find(&o).Error
	return
}

func (r AssetRepository) FindByIds(assetIds []string) (o []models.Asset, err error) {
	err = r.DB.Where("id in ?", assetIds).Find(&o).Error
	return
}

func (r AssetRepository) FindByProtocol(protocol string) (o []models.Asset, err error) {
	err = r.DB.Where("protocol = ?", protocol).Find(&o).Error
	return
}

func (r AssetRepository) FindByProtocolAndIds(protocol string, assetIds []string) (o []models.Asset, err error) {
	err = r.DB.Where("protocol = ? and id in ?", protocol, assetIds).Find(&o).Error
	return
}

func (r AssetRepository) FindByProtocolAndUser(protocol string, account models.User) (o []models.Asset, err error) {
	db := r.DB.Table("assets").Select("assets.id,assets.name,assets.ip,assets.port,assets.protocol,assets.active,assets.owner,assets.created_at, users.nickname as owner_name,COUNT(resource_sharers.user_id) as sharer_count").Joins("left join users on assets.owner = users.id").Joins("left join resource_sharers on assets.id = resource_sharers.resource_id").Group("assets.id")

	if constants.RoleDefault == account.Role {
		owner := account.ID
		db = db.Where("assets.owner = ? or resource_sharers.user_id = ?", owner, owner)
	}

	if len(protocol) > 0 {
		db = db.Where("assets.protocol = ?", protocol)
	}
	err = db.Find(&o).Error
	return
}

func (r AssetRepository) Find(pageIndex, pageSize int, name, protocol, tags string, account models.User, owner, sharer, userGroupId, ip, order, field string) (o []models.AssetForPage, total int64, err error) {
	db := r.DB.Table("assets").Select("assets.id,assets.name,assets.ip,assets.port,assets.protocol,assets.active,assets.owner,assets.created_at,assets.tags, users.nickname as owner_name,COUNT(resource_sharers.user_id) as sharer_count").Joins("left join users on assets.owner = users.id").Joins("left join resource_sharers on assets.id = resource_sharers.resource_id").Group("assets.id")
	dbCounter := r.DB.Table("assets").Select("DISTINCT assets.id").Joins("left join resource_sharers on assets.id = resource_sharers.resource_id").Group("assets.id")

	if constants.RoleDefault == account.Role {
		owner := account.ID
		db = db.Where("assets.owner = ? or resource_sharers.user_id = ?", owner, owner)
		dbCounter = dbCounter.Where("assets.owner = ? or resource_sharers.user_id = ?", owner, owner)

		// 查询用户所在用户组列表
		userGroupIds, err := userGroupRepository.FindUserGroupIdsByUserId(account.ID)
		if err != nil {
			return nil, 0, err
		}

		if len(userGroupIds) > 0 {
			db = db.Or("resource_sharers.user_group_id in ?", userGroupIds)
			dbCounter = dbCounter.Or("resource_sharers.user_group_id in ?", userGroupIds)
		}
	} else {
		if len(owner) > 0 {
			db = db.Where("assets.owner = ?", owner)
			dbCounter = dbCounter.Where("assets.owner = ?", owner)
		}
		if len(sharer) > 0 {
			db = db.Where("resource_sharers.user_id = ?", sharer)
			dbCounter = dbCounter.Where("resource_sharers.user_id = ?", sharer)
		}

		if len(userGroupId) > 0 {
			db = db.Where("resource_sharers.user_group_id = ?", userGroupId)
			dbCounter = dbCounter.Where("resource_sharers.user_group_id = ?", userGroupId)
		}
	}

	if len(name) > 0 {
		db = db.Where("assets.name like ?", "%"+name+"%")
		dbCounter = dbCounter.Where("assets.name like ?", "%"+name+"%")
	}

	if len(ip) > 0 {
		db = db.Where("assets.ip like ?", "%"+ip+"%")
		dbCounter = dbCounter.Where("assets.ip like ?", "%"+ip+"%")
	}

	if len(protocol) > 0 {
		db = db.Where("assets.protocol = ?", protocol)
		dbCounter = dbCounter.Where("assets.protocol = ?", protocol)
	}

	if len(tags) > 0 {
		tagArr := strings.Split(tags, ",")
		for i := range tagArr {
			if viper.GetString("db.type") == "sqlite" {
				db = db.Where("(',' || assets.tags || ',') LIKE ?", "%,"+tagArr[i]+",%")
				dbCounter = dbCounter.Where("(',' || assets.tags || ',') LIKE ?", "%,"+tagArr[i]+",%")
			} else {
				db = db.Where("find_in_set(?, assets.tags)", tagArr[i])
				dbCounter = dbCounter.Where("find_in_set(?, assets.tags)", tagArr[i])
			}
		}
	}

	err = dbCounter.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if order == "ascend" {
		order = "asc"
	} else {
		order = "desc"
	}

	if field == "name" {
		field = "name"
	} else {
		field = "created_at"
	}

	err = db.Order("assets." + field + " " + order).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&o).Error

	if o == nil {
		o = make([]models.AssetForPage, 0)
	} else {
		for i := 0; i < len(o); i++ {
			if o[i].Protocol == "ssh" {
				attributes, err := r.FindAttrById(o[i].ID)
				if err != nil {
					continue
				}

				for j := range attributes {
					if attributes[j].Name == constants.SshMode {
						o[i].SshMode = attributes[j].Value
						break
					}
				}
			}
		}
	}
	return
}

func (r AssetRepository) Encrypt(item *models.Asset, password []byte) error {
	if item.Password != "" && item.Password != "-" {
		encryptedCBC, err := utils.AesEncryptCBC([]byte(item.Password), password)
		if err != nil {
			return err
		}
		item.Password = base64.StdEncoding.EncodeToString(encryptedCBC)
	}
	if item.PrivateKey != "" && item.PrivateKey != "-" {
		encryptedCBC, err := utils.AesEncryptCBC([]byte(item.PrivateKey), password)
		if err != nil {
			return err
		}
		item.PrivateKey = base64.StdEncoding.EncodeToString(encryptedCBC)
	}
	if item.Passphrase != "" && item.Passphrase != "-" {
		encryptedCBC, err := utils.AesEncryptCBC([]byte(item.Passphrase), password)
		if err != nil {
			return err
		}
		item.Passphrase = base64.StdEncoding.EncodeToString(encryptedCBC)
	}
	item.Encrypted = true
	return nil
}

func (r AssetRepository) Create(o *models.Asset) (err error) {
	if err := r.Encrypt(o, utils.Encryption()); err != nil {
		return err
	}
	if err = r.DB.Create(o).Error; err != nil {
		return err
	}
	return nil
}

func (r AssetRepository) FindById(id string) (o models.Asset, err error) {
	err = r.DB.Where("id = ?", id).First(&o).Error
	return
}

func (r AssetRepository) Decrypt(item *models.Asset, password []byte) error {
	if item.Encrypted {
		if item.Password != "" && item.Password != "-" {
			origData, err := base64.StdEncoding.DecodeString(item.Password)
			if err != nil {
				return err
			}
			decryptedCBC, err := utils.AesDecryptCBC(origData, password)
			if err != nil {
				return err
			}
			item.Password = string(decryptedCBC)
		}
		if item.PrivateKey != "" && item.PrivateKey != "-" {
			origData, err := base64.StdEncoding.DecodeString(item.PrivateKey)
			if err != nil {
				return err
			}
			decryptedCBC, err := utils.AesDecryptCBC(origData, password)
			if err != nil {
				return err
			}
			item.PrivateKey = string(decryptedCBC)
		}
		if item.Passphrase != "" && item.Passphrase != "-" {
			origData, err := base64.StdEncoding.DecodeString(item.Passphrase)
			if err != nil {
				return err
			}
			decryptedCBC, err := utils.AesDecryptCBC(origData, password)
			if err != nil {
				return err
			}
			item.Passphrase = string(decryptedCBC)
		}
	}
	return nil
}

func (r AssetRepository) FindByIdAndDecrypt(id string) (o models.Asset, err error) {
	err = r.DB.Where("id = ?", id).First(&o).Error
	if err == nil {
		err = r.Decrypt(&o, utils.Encryption())
	}
	return
}

func (r AssetRepository) UpdateById(o *models.Asset, id string) error {
	o.ID = id
	return r.DB.Updates(o).Error
}

func (r AssetRepository) UpdateActiveById(active bool, id string) error {
	sql := "update assets set active = ? where id = ?"
	return r.DB.Exec(sql, active, id).Error
}

func (r AssetRepository) DeleteById(id string) error {
	return r.DB.Where("id = ?", id).Delete(&models.Asset{}).Error
}

func (r AssetRepository) Count() (total int64, err error) {
	err = r.DB.Find(&models.Asset{}).Count(&total).Error
	return
}

func (r AssetRepository) CountByUserId(userId string) (total int64, err error) {
	db := r.DB.Joins("left join resource_sharers on assets.id = resource_sharers.resource_id")

	db = db.Where("assets.owner = ? or resource_sharers.user_id = ?", userId, userId)

	// 查询用户所在用户组列表
	userGroupIds, err := userGroupRepository.FindUserGroupIdsByUserId(userId)
	if err != nil {
		return 0, err
	}

	if len(userGroupIds) > 0 {
		db = db.Or("resource_sharers.user_group_id in ?", userGroupIds)
	}
	err = db.Find(&models.Asset{}).Count(&total).Error
	return
}

func (r AssetRepository) FindTags() (o []string, err error) {
	var assets []models.Asset
	err = r.DB.Not("tags = ?", "").Find(&assets).Error
	if err != nil {
		return nil, err
	}

	o = make([]string, 0)

	for i := range assets {
		if len(assets[i].Tags) == 0 {
			continue
		}
		split := strings.Split(assets[i].Tags, ",")

		o = append(o, split...)
	}

	return utils.Distinct(o), nil
}

func (r AssetRepository) UpdateAttributes(assetId, protocol string, m map[string]interface{}) error {
	var data []models.AssetAttribute
	var parameterNames []string
	switch protocol {
	case "ssh":
		parameterNames = constants.SSHParameterNames
	case "rdp":
		parameterNames = constants.RDPParameterNames
	case "vnc":
		parameterNames = constants.VNCParameterNames
	case "telnet":
		parameterNames = constants.TelnetParameterNames
	case "kubernetes":
		parameterNames = constants.KubernetesParameterNames
	}

	for i := range parameterNames {
		name := parameterNames[i]
		if m[name] != nil && m[name] != "" {
			data = append(data, genAttribute(assetId, name, m))
		}
	}

	return r.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("asset_id = ?", assetId).Delete(&models.AssetAttribute{}).Error
		if err != nil {
			return err
		}
		return tx.CreateInBatches(&data, len(data)).Error
	})
}

func genAttribute(assetId, name string, m map[string]interface{}) models.AssetAttribute {
	value := fmt.Sprintf("%v", m[name])
	attribute := models.AssetAttribute{
		Id:      utils.Sign([]string{assetId, name}),
		AssetId: assetId,
		Name:    name,
		Value:   value,
	}
	return attribute
}

func (r AssetRepository) FindAttrById(assetId string) (o []models.AssetAttribute, err error) {
	err = r.DB.Where("asset_id = ?", assetId).Find(&o).Error
	if o == nil {
		o = make([]models.AssetAttribute, 0)
	}
	return o, err
}

func (r AssetRepository) FindAssetAttrMapByAssetId(assetId string) (map[string]interface{}, error) {
	asset, err := r.FindById(assetId)
	if err != nil {
		return nil, err
	}
	attributes, err := r.FindAttrById(assetId)
	if err != nil {
		return nil, err
	}

	var parameterNames []string
	switch asset.Protocol {
	case "ssh":
		parameterNames = constants.SSHParameterNames
	case "rdp":
		parameterNames = constants.RDPParameterNames
	case "vnc":
		parameterNames = constants.VNCParameterNames
	case "telnet":
		parameterNames = constants.TelnetParameterNames
	case "kubernetes":
		parameterNames = constants.KubernetesParameterNames
	}
	// TODO nil panic
	cfgsMap := configsRepository.FindAllMap()
	var attributeMap = make(map[string]interface{})
	for name := range cfgsMap {
		if utils.Contains(parameterNames, name) {
			attributeMap[name] = cfgsMap[name]
		}
	}

	for i := range attributes {
		attributeMap[attributes[i].Name] = attributes[i].Value
	}
	return attributeMap, nil
}

func (r AssetRepository) InitAsset(item *models.Asset, m map[string]interface{}) error {
	// 已存在资产跳过
	exist, _ := r.AssetGet("ip = ?", item.IP)
	if exist != nil {
		zlog.Info("已存在资产 %v 忽略", item.IP)
		return nil
	}
	if err := r.Create(item); err != nil {
		return err
	}

	if err := r.UpdateAttributes(item.ID, item.Protocol, m); err != nil {
		return err
	}

	// 创建后自动检测资产是否存活
	go func() {
		active := utils.Tcping(item.IP, item.Port)
		_ = r.UpdateActiveById(active, item.ID)
	}()
	return nil
}

func (r AssetRepository) AssetGet(where string, args ...interface{}) (*models.Asset, error) {
	var u models.Asset
	err := r.DB.Model(models.Asset{}).Where(where, args...).Last(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}
