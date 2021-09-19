package repository

import (
	"gorm.io/gorm"
	"next-terminal/models"
	"next-terminal/pkg/guacd"
)

type ConfigsRepository struct {
	DB *gorm.DB
}

func NewConfigsRepository(db *gorm.DB) *ConfigsRepository {
	configsRepository = &ConfigsRepository{DB: db}
	return configsRepository
}

func (r ConfigsRepository) FindAll() (o []models.Configs) {
	if r.DB.Find(&o).Error != nil {
		return nil
	}
	return
}

func (r ConfigsRepository) Create(o *models.Configs) (err error) {
	err = r.DB.Create(o).Error
	return
}

func (c ConfigsRepository) Save(o *models.Configs) error {
	return c.DB.Where("ckey = ?", o.Ckey).Save(o).Error
}

func (r ConfigsRepository) UpdateByName(o *models.Configs, ckey string) error {
	o.Ckey = ckey
	return r.DB.Updates(o).Error
}

func (r ConfigsRepository) FindByName(ckey string) (o models.Configs, err error) {
	err = r.DB.Where("ckey = ?", ckey).First(&o).Error
	return
}

func (r ConfigsRepository) FindAllMap() map[string]string {
	cfgs := r.FindAll()
	cfgMap := make(map[string]string)
	for i := range cfgs {
		cfgMap[cfgs[i].Ckey] = cfgs[i].Cval
	}
	return cfgMap
}

func (r ConfigsRepository) GetDrivePath() (string, error) {
	cfg, err := r.FindByName(guacd.DrivePath)
	if err != nil {
		return "", err
	}
	return cfg.Cval, nil
}

func (r ConfigsRepository) GetRecordingPath() (string, error) {
	cfg, err := r.FindByName(guacd.RecordingPath)
	if err != nil {
		return "", err
	}
	return cfg.Cval, nil
}

// ConfigsGet 获取配置
func (r ConfigsRepository) ConfigsGet(ckey string) (string, error) {
	var obj models.Configs
	has := r.DB.Model(models.Configs{}).Where("ckey=?", ckey).Last(&obj)
	if has.Error != nil && has.Error != gorm.ErrRecordNotFound {
		return "", has.Error
	}

	if has.RowsAffected == 0 {
		return "", nil
	}

	return obj.Cval, nil
}

// ConfigsSet 添加配置
func (r ConfigsRepository) ConfigsSet(ckey, cval string) error {
	var obj models.Configs
	has := r.DB.Model(models.Configs{}).Where("ckey=?", ckey).Last(&obj)
	if has.Error != nil && has.Error != gorm.ErrRecordNotFound {
		return has.Error
	}
	var err error
	if has.RowsAffected == 0 {
		err = r.Create(&models.Configs{
			Ckey: ckey,
			Cval: cval,
		})
	} else {
		obj.Cval = cval
		err = r.Save(&obj)
	}
	return err
}
