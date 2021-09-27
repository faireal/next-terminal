package repository

import (
	"gorm.io/gorm"
	"next-terminal/models"
)

type LogsRepository struct {
	DB *gorm.DB
}

func NewLogsRepository(db *gorm.DB) *LogsRepository {
	logsRepository = &LogsRepository{DB: db}
	return logsRepository
}

func (r LogsRepository) Find(pageIndex, pageSize int, userId, clientIp string) (o []models.LogsForPage, total int64, err error) {

	db := r.DB.Table("logs").Select("logs.*, users.nickname as user_name").Joins("left join users on logs.user_id = users.id")
	dbCounter := r.DB.Table("logs").Select("DISTINCT logs.id")

	if userId != "" {
		db = db.Where("logs.user_id = ?", userId)
		dbCounter = dbCounter.Where("logs.user_id = ?", userId)
	}

	if clientIp != "" {
		db = db.Where("logs.client_ip like ?", "%"+clientIp+"%")
		dbCounter = dbCounter.Where("logs.client_ip like ?", "%"+clientIp+"%")
	}

	err = dbCounter.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Order("logs.login_time desc").Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&o).Error
	if o == nil {
		o = make([]models.LogsForPage, 0)
	}
	return
}

func (r LogsRepository) FindAliveLoginLogs() (o []models.Logs, err error) {
	err = r.DB.Where("logout_time is null").Find(&o).Error
	return
}

func (r LogsRepository) FindAliveLoginLogsByUserId(userId string) (o []models.Logs, err error) {
	err = r.DB.Where("logout_time is null and user_id = ?", userId).Find(&o).Error
	return
}

func (r LogsRepository) Create(o *models.Logs) (err error) {
	return r.DB.Create(o).Error
}

func (r LogsRepository) DeleteByIdIn(ids []string) (err error) {
	return r.DB.Where("id in ?", ids).Delete(&models.Logs{}).Error
}

func (r LogsRepository) FindById(id string) (o models.Logs, err error) {
	err = r.DB.Where("id = ?", id).First(&o).Error
	return
}

func (r LogsRepository) Update(o *models.Logs) error {
	return r.DB.Updates(o).Error
}
