package models

type Job struct {
	Model
	ID          string         `gorm:"primary_key" json:"id"`
	CronJobId   int            `json:"cronJobId"`
	Name        string         `json:"name"`
	Func        string         `json:"func"`
	Cron        string         `json:"cron"`
	Mode        string         `json:"mode"`
	ResourceIds string         `json:"resourceIds"`
	Status      string         `json:"status"`
	Metadata    string         `json:"metadata"`
}

func (r *Job) TableName() string {
	return "jobs"
}

type JobLog struct {
	Model
	ID        string         `json:"id"`
	JobId     string         `json:"jobId"`
	Message   string         `json:"message"`
}

func (r *JobLog) TableName() string {
	return "job_logs"
}

func init() {
	Migrate(Job{})
	Migrate(JobLog{})
}
