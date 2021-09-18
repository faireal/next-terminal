package repository

import (
	"fmt"
	"gorm.io/gorm"
	"next-terminal/models"
	"next-terminal/pkg/utils"
)

type ResourceSharerRepository struct {
	DB *gorm.DB
}

func NewResourceSharerRepository(db *gorm.DB) *ResourceSharerRepository {
	resourceSharerRepository = &ResourceSharerRepository{DB: db}
	return resourceSharerRepository
}

func (r *ResourceSharerRepository) FindUserIdsByResourceId(resourceId string) (o []string, err error) {
	err = r.DB.Table("resource_sharers").Select("user_id").Where("resource_id = ?", resourceId).Find(&o).Error
	if o == nil {
		o = make([]string, 0)
	}
	return
}

func (r *ResourceSharerRepository) OverwriteUserIdsByResourceId(resourceId, resourceType string, userIds []string) (err error) {
	db := r.DB.Begin()

	var owner string
	// 检查资产是否存在
	switch resourceType {
	case "asset":
		resource := models.Asset{}
		err = db.Where("id = ?", resourceId).First(&resource).Error
		owner = resource.Owner
	case "command":
		resource := models.Command{}
		err = db.Where("id = ?", resourceId).First(&resource).Error
		owner = resource.Owner
	case "credential":
		resource := models.Credential{}
		err = db.Where("id = ?", resourceId).First(&resource).Error
		owner = resource.Owner
	}

	if err == gorm.ErrRecordNotFound {
		return fmt.Errorf("资源「%v」不存在", resourceId)
	}

	for i := range userIds {
		if owner == userIds[i] {
			return fmt.Errorf("参数错误")
		}
	}

	db.Where("resource_id = ?", resourceId).Delete(&models.ResourceSharer{})

	for i := range userIds {
		userId := userIds[i]
		if len(userId) == 0 {
			continue
		}
		id := utils.Sign([]string{resourceId, resourceType, userId})
		resource := &models.ResourceSharer{
			ID:           id,
			ResourceId:   resourceId,
			ResourceType: resourceType,
			UserId:       userId,
		}
		err = db.Create(resource).Error
		if err != nil {
			return err
		}
	}
	db.Commit()
	return nil
}

func (r *ResourceSharerRepository) DeleteByUserIdAndResourceTypeAndResourceIdIn(userGroupId, userId, resourceType string, resourceIds []string) error {
	db := r.DB
	if userGroupId != "" {
		db = db.Where("user_group_id = ?", userGroupId)
	}

	if userId != "" {
		db = db.Where("user_id = ?", userId)
	}

	if resourceType != "" {
		db = db.Where("resource_type = ?", resourceType)
	}

	if len(resourceIds) > 0 {
		db = db.Where("resource_id in ?", resourceIds)
	}

	return db.Delete(&models.ResourceSharer{}).Error
}

func (r *ResourceSharerRepository) DeleteResourceSharerByResourceId(resourceId string) error {
	return r.DB.Where("resource_id = ?", resourceId).Delete(&models.ResourceSharer{}).Error
}

func (r *ResourceSharerRepository) AddSharerResources(userGroupId, userId, resourceType string, resourceIds []string) error {
	return r.DB.Transaction(func(tx *gorm.DB) (err error) {

		for i := range resourceIds {
			resourceId := resourceIds[i]

			var owner string
			// 检查资产是否存在
			switch resourceType {
			case "asset":
				resource := models.Asset{}
				if err = tx.Where("id = ?", resourceId).First(&resource).Error; err != nil {
					return fmt.Errorf("find asset fail err: %v", err)
				}
				owner = resource.Owner
			case "command":
				resource := models.Command{}
				if err = tx.Where("id = ?", resourceId).First(&resource).Error; err != nil {
					return fmt.Errorf("find command fail err: %v", err)
				}
				owner = resource.Owner
			case "credential":
				resource := models.Credential{}
				if err = tx.Where("id = ?", resourceId).First(&resource).Error; err != nil {
					return fmt.Errorf("find credential fail err: %v", err)

				}
				owner = resource.Owner
			}

			if owner == userId {
				return fmt.Errorf("参数错误")
			}

			id := utils.Sign([]string{resourceId, resourceType, userId, userGroupId})
			resource := &models.ResourceSharer{
				ID:           id,
				ResourceId:   resourceId,
				ResourceType: resourceType,
				UserId:       userId,
				UserGroupId:  userGroupId,
			}
			err = tx.Create(resource).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ResourceSharerRepository) FindAssetIdsByUserId(userId string) (assetIds []string, err error) {
	// 查询当前用户创建的资产
	var ownerAssetIds, sharerAssetIds []string
	asset := models.Asset{}
	err = r.DB.Table(asset.TableName()).Select("id").Where("owner = ?", userId).Find(&ownerAssetIds).Error
	if err != nil {
		return nil, err
	}

	// 查询其他用户授权给该用户的资产
	groupIds, err := userGroupRepository.FindUserGroupIdsByUserId(userId)
	if err != nil {
		return nil, err
	}

	db := r.DB.Table("resource_sharers").Select("resource_id").Where("user_id = ?", userId)
	if len(groupIds) > 0 {
		db = db.Or("user_group_id in ?", groupIds)
	}
	err = db.Find(&sharerAssetIds).Error
	if err != nil {
		return nil, err
	}

	// 合并查询到的资产ID
	assetIds = make([]string, 0)

	if ownerAssetIds != nil {
		assetIds = append(assetIds, ownerAssetIds...)
	}

	if sharerAssetIds != nil {
		assetIds = append(assetIds, sharerAssetIds...)
	}

	return
}
