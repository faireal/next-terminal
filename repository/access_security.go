package repository

import (
	"gorm.io/gorm"
	"next-terminal/models"
)

type AccessSecurityRepository struct {
	DB *gorm.DB
}

func NewAccessSecurityRepository(db *gorm.DB) *AccessSecurityRepository {
	accessSecurityRepository = &AccessSecurityRepository{DB: db}
	return accessSecurityRepository
}

func (r AccessSecurityRepository) FindAllAccessSecurities() (o []models.AccessSecurity, err error) {
	db := r.DB
	err = db.Order("priority asc").Find(&o).Error
	return
}

func (r AccessSecurityRepository) Find(pageIndex, pageSize int, ip, rule, order, field string) (o []models.AccessSecurity, total int64, err error) {
	t := models.AccessSecurity{}
	db := r.DB.Table(t.TableName())
	dbCounter := r.DB.Table(t.TableName())

	if len(ip) > 0 {
		db = db.Where("ip like ?", "%"+ip+"%")
		dbCounter = dbCounter.Where("ip like ?", "%"+ip+"%")
	}

	if len(rule) > 0 {
		db = db.Where("rule = ?", rule)
		dbCounter = dbCounter.Where("rule = ?", rule)
	}

	err = dbCounter.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if order == "descend" {
		order = "desc"
	} else {
		order = "asc"
	}

	if field == "ip" {
		field = "ip"
	} else if field == "rule" {
		field = "rule"
	} else {
		field = "priority"
	}

	err = db.Order(field + " " + order).Find(&o).Offset((pageIndex - 1) * pageSize).Limit(pageSize).Error
	if o == nil {
		o = make([]models.AccessSecurity, 0)
	}
	return
}

func (r AccessSecurityRepository) Create(o *models.AccessSecurity) error {
	return r.DB.Create(o).Error
}

func (r AccessSecurityRepository) UpdateByID(o *models.AccessSecurity, id string) error {
	o.ID = id
	return r.DB.Updates(o).Error
}

func (r AccessSecurityRepository) DeleteByID(id string) error {
	return r.DB.Where("id = ?", id).Delete(models.AccessSecurity{}).Error
}

func (r AccessSecurityRepository) FindByID(id string) (o *models.AccessSecurity, err error) {
	err = r.DB.Where("id = ?", id).First(&o).Error
	return
}
