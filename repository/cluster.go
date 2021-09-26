// AGPL License
// Copyright (c) 2021 ysicing <i@ysicing.me>

package repository

import (
	"encoding/base64"
	"github.com/ergoapi/zlog"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"next-terminal/constants"
	"next-terminal/models"
	"next-terminal/pkg/utils"
	"strings"
)

type ClusterRepository struct {
	DB *gorm.DB
}

func NewClusterRepository(db *gorm.DB) *ClusterRepository {
	clusterRepository = &ClusterRepository{DB: db}
	return clusterRepository
}

func (c ClusterRepository) FindAll() (o []models.Cloud, err error) {
	err = c.DB.Find(&o).Error
	return
}

func (c ClusterRepository) FindByProtocolAndUser(account models.User) (o []models.Cluster, err error) {
	db := c.DB.Model(&models.Cluster{}).Select("cluster.*, users.nickname as owner_name").Joins("left join users on cluster.owner = users.id").Group("cluster.id")

	if constants.RoleDefault == account.Role {
		owner := account.ID
		db = db.Where("cluster.owner = ?", owner)
	}

	err = db.Find(&o).Error
	return
}

func (c ClusterRepository) Find(pageIndex, pageSize int, name, tags string, account models.User, owner, order, field string) (o []models.Cluster, total int64, err error) {
	db := c.DB.Table("clusters").Select("clusters.*, users.nickname as owner_name").Joins("left join users on clusters.owner = users.id").Group("clusters.id")
	dbCounter := c.DB.Table("clusters").Select("DISTINCT clusters.id").Group("clusters.id")

	if constants.RoleDefault == account.Role {
		owner := account.ID
		db = db.Where("clusters.owner = ?", owner)
		dbCounter = dbCounter.Where("clusters.owner = ?", owner)

		// 查询用户所在用户组列表
		// TODO
	} else {
		if len(owner) > 0 {
			db = db.Where("clusters.owner = ?", owner)
			dbCounter = dbCounter.Where("clusters.owner = ?", owner)
		}
	}

	if len(name) > 0 {
		db = db.Where("clusters.name like ?", "%"+name+"%")
		dbCounter = dbCounter.Where("clusters.name like ?", "%"+name+"%")
	}

	if len(tags) > 0 {
		tagArr := strings.Split(tags, ",")
		for i := range tagArr {
			if viper.GetString("db.type") == "sqlite" {
				db = db.Where("(',' || clusters.tags || ',') LIKE ?", "%,"+tagArr[i]+",%")
				dbCounter = dbCounter.Where("(',' || clusters.tags || ',') LIKE ?", "%,"+tagArr[i]+",%")
			} else {
				db = db.Where("find_in_set(?, clusters.tags)", tagArr[i])
				dbCounter = dbCounter.Where("find_in_set(?, clusters.tags)", tagArr[i])
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

	err = db.Order("clusters." + field + " " + order).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&o).Error

	if o == nil {
		o = make([]models.Cluster, 0)
	}
	return
}

func (c ClusterRepository) Encrypt(item *models.Cluster, password []byte) error {
	if item.Kubeconfig != "" {
		encryptedCBC, err := utils.AesEncryptCBC([]byte(item.Kubeconfig), password)
		if err != nil {
			return err
		}
		item.Kubeconfig = base64.StdEncoding.EncodeToString(encryptedCBC)
	}
	item.Encrypted = true
	return nil
}

func (c ClusterRepository) Decrypt(item *models.Cluster, password []byte) error {
	if item.Encrypted {
		if item.Kubeconfig != "" {
			origData, err := base64.StdEncoding.DecodeString(item.Kubeconfig)
			if err != nil {
				return err
			}
			decryptedCBC, err := utils.AesDecryptCBC(origData, password)
			if err != nil {
				return err
			}
			item.Kubeconfig = string(decryptedCBC)
		}
	}
	return nil
}

func (c ClusterRepository) Create(o *models.Cluster) (err error) {
	if err := c.Encrypt(o, utils.Encryption()); err != nil {
		return err
	}
	if err = c.DB.Create(o).Error; err != nil {
		return err
	}
	return nil
}

func (c ClusterRepository) UpdateStatusByID(active bool, id string) error {
	sql := "update clusters set status = ? where id = ?"
	return c.DB.Exec(sql, active, id).Error
}

func (c ClusterRepository) InitCluster(item *models.Cluster, m map[string]interface{}) error {
	// TODO已存在集群跳过
	exist, _ := c.Get("name = ?", item.Name)
	if exist != nil {
		zlog.Info("已存在集群 %v 忽略", item.Name)
		return nil
	}
	if err := c.Create(item); err != nil {
		return err
	}

	// 创建后检测集群是否可用
	go func() {
		active := false // utils.Tcping(item.IP, item.Port)
		_ = c.UpdateStatusByID(active, item.ID)
	}()
	return nil
}

func (c ClusterRepository) Get(where string, args ...interface{}) (*models.Cluster, error) {
	var u models.Cluster
	err := c.DB.Model(models.Cluster{}).Where(where, args...).Last(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}
