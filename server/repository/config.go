package repository

import (
	"next-terminal/pkg/guacd"
	"next-terminal/server/model"

	"gorm.io/gorm"
)

type ConfigsRepository struct {
	DB *gorm.DB
}

func NewConfigsRepository(db *gorm.DB) *ConfigsRepository {
	configsRepository = &ConfigsRepository{DB: db}
	return configsRepository
}

func (r ConfigsRepository) FindAll() (o []model.Configs) {
	if r.DB.Find(&o).Error != nil {
		return nil
	}
	return
}

func (r ConfigsRepository) Create(o *model.Configs) (err error) {
	err = r.DB.Create(o).Error
	return
}

func (r ConfigsRepository) UpdateByName(o *model.Configs, ckey string) error {
	o.Ckey = ckey
	return r.DB.Updates(o).Error
}

func (r ConfigsRepository) FindByName(ckey string) (o model.Configs, err error) {
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
