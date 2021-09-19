package repository

import (
	"gorm.io/gorm"
	"next-terminal/models"
)

type JobLogRepository struct {
	DB *gorm.DB
}

func NewJobLogRepository(db *gorm.DB) *JobLogRepository {
	jobLogRepository = &JobLogRepository{DB: db}
	return jobLogRepository
}

func (r JobLogRepository) Create(o *models.JobLog) error {
	return r.DB.Create(o).Error
}

func (r JobLogRepository) FindByJobId(jobId string) (o []models.JobLog, err error) {
	err = r.DB.Where("job_id = ?", jobId).Order("created_at asc").Find(&o).Error
	return
}

func (r JobLogRepository) DeleteByJobId(jobId string) error {
	return r.DB.Where("job_id = ?", jobId).Delete(models.JobLog{}).Error
}
