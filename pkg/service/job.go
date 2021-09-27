package service

import (
	"encoding/json"
	"fmt"
	"github.com/ergoapi/util/zos"
	"next-terminal/constants"
	"next-terminal/models"
	"next-terminal/pkg/utils"
	"next-terminal/repository"
	"strings"
	"time"

	"github.com/ergoapi/zlog"
	"github.com/robfig/cron/v3"
	cronv1 "next-terminal/pkg/cron"
	"next-terminal/pkg/term"
)

type JobService struct {
	jobRepository        *repository.JobRepository
	jobLogRepository     *repository.JobLogRepository
	assetRepository      *repository.AssetRepository
	clusterRepository    *repository.ClusterRepository
	credentialRepository *repository.CredentialRepository
}

func NewJobService(jobRepository *repository.JobRepository, jobLogRepository *repository.JobLogRepository,
	assetRepository *repository.AssetRepository, clusterRepository *repository.ClusterRepository,
	credentialRepository *repository.CredentialRepository) *JobService {
	return &JobService{jobRepository: jobRepository, jobLogRepository: jobLogRepository,
		assetRepository: assetRepository, clusterRepository: clusterRepository,
		credentialRepository: credentialRepository}
}

func (r JobService) ChangeStatusById(id, status string) error {
	job, err := r.jobRepository.FindById(id)
	if err != nil {
		return err
	}
	if status == constants.JobStatusRunning {
		j, err := getJob(&job, &r)
		if err != nil {
			return err
		}
		entryID, err := cronv1.Cron.AddJob(job.Cron, j)
		if err != nil {
			return err
		}
		zlog.Debug("开启计划任务「%v」,运行中计划任务数量「%v」", job.Name, len(cronv1.Cron.Entries()))

		jobForUpdate := models.Job{ID: id, Status: constants.JobStatusRunning, CronJobId: int(entryID)}

		return r.jobRepository.UpdateById(&jobForUpdate)
	} else {
		cronv1.Cron.Remove(cron.EntryID(job.CronJobId))
		zlog.Debug("关闭计划任务「%v」,运行中计划任务数量「%v」", job.Name, len(cronv1.Cron.Entries()))
		jobForUpdate := models.Job{ID: id, Status: constants.JobStatusNotRunning}
		return r.jobRepository.UpdateById(&jobForUpdate)
	}
}

func getJob(j *models.Job, jobService *JobService) (job cron.Job, err error) {
	switch j.Func {
	case constants.FuncCheckAssetStatusJob:
		job = CheckAssetStatusJob{ID: j.ID, Mode: j.Mode, ResourceIds: j.ResourceIds, Metadata: j.Metadata, jobService: jobService}
	case constants.FuncCheckClusterStatusJob:
		job = CheckClusterStatusJob{ID: j.ID, Mode: j.Mode, ResourceIds: j.ResourceIds, Metadata: j.Metadata, jobService: jobService}
	case constants.FuncShellJob:
		job = ShellJob{ID: j.ID, Mode: j.Mode, ResourceIds: j.ResourceIds, Metadata: j.Metadata, jobService: jobService}
	default:
		return nil, fmt.Errorf("未识别的任务")
	}
	return job, err
}

type CheckAssetStatusJob struct {
	ID          string
	Mode        string
	ResourceIds string
	Metadata    string
	jobService  *JobService
}

func (r CheckAssetStatusJob) Run() {
	if r.ID == "" {
		return
	}

	var assets []models.Asset
	if r.Mode == constants.JobModeAll {
		assets, _ = r.jobService.assetRepository.FindAll()
	} else {
		assets, _ = r.jobService.assetRepository.FindByIds(strings.Split(r.ResourceIds, ","))
	}

	if len(assets) == 0 {
		return
	}

	msgChan := make(chan string)
	for i := range assets {
		asset := assets[i]
		go func() {
			t1 := time.Now()
			active := utils.Tcping(asset.IP, asset.Port)
			elapsed := time.Since(t1)
			msg := fmt.Sprintf("资产「%v」存活状态检测完成，存活「%v」，耗时「%v」", asset.Name, active, elapsed)

			_ = r.jobService.assetRepository.UpdateActiveById(active, asset.ID)
			zlog.Info(msg)
			msgChan <- msg
		}()
	}

	var message = ""
	for i := 0; i < len(assets); i++ {
		message += <-msgChan + "\n"
	}

	_ = r.jobService.jobRepository.UpdateLastUpdatedById(r.ID)
	jobLog := models.JobLog{
		ID:      zos.GenUUID(),
		JobId:   r.ID,
		Message: message,
	}

	_ = r.jobService.jobLogRepository.Create(&jobLog)
}

type CheckClusterStatusJob struct {
	ID          string
	Mode        string
	ResourceIds string
	Metadata    string
	jobService  *JobService
}

func (r CheckClusterStatusJob) Run() {
	if r.ID == "" {
		return
	}

	var clusters []models.Cluster
	if r.Mode == constants.JobModeAll {
		clusters, _ = r.jobService.clusterRepository.Gets("")
	} else {
		clusters, _ = r.jobService.clusterRepository.FindByID(strings.Split(r.ResourceIds, ","))
	}

	if len(clusters) == 0 {
		return
	}

	msgChan := make(chan string)
	for i := range clusters {
		cluster := clusters[i]
		go func() {
			t1 := time.Now()
			active := utils.KubePing(cluster.Kubeconfig, cluster.ID)
			elapsed := time.Since(t1)
			msg := fmt.Sprintf("集群「%v(%v)」存活状态检测完成，存活「%v」，耗时「%v」", cluster.Name, cluster.ID, active, elapsed)
			_ = r.jobService.clusterRepository.UpdateStatusByID(active, cluster.ID)
			zlog.Info(msg)
			msgChan <- msg
		}()
	}

	var message = ""
	for i := 0; i < len(clusters); i++ {
		message += <-msgChan + "\n"
	}
	_ = r.jobService.jobRepository.UpdateLastUpdatedById(r.ID)
	jobLog := models.JobLog{
		ID:      zos.GenUUID(),
		JobId:   r.ID,
		Message: message,
	}
	_ = r.jobService.jobLogRepository.Create(&jobLog)
}

type ShellJob struct {
	ID          string
	Mode        string
	ResourceIds string
	Metadata    string
	jobService  *JobService
}

type MetadataShell struct {
	Shell string
}

func (r ShellJob) Run() {
	if r.ID == "" {
		return
	}

	var assets []models.Asset
	if r.Mode == constants.JobModeAll {
		assets, _ = r.jobService.assetRepository.FindByProtocol("ssh")
	} else {
		assets, _ = r.jobService.assetRepository.FindByProtocolAndIds("ssh", strings.Split(r.ResourceIds, ","))
	}

	if len(assets) == 0 {
		return
	}

	var metadataShell MetadataShell
	err := json.Unmarshal([]byte(r.Metadata), &metadataShell)
	if err != nil {
		zlog.Error("JSON数据解析失败 %v", err)
		return
	}

	msgChan := make(chan string)
	for i := range assets {
		asset, err := r.jobService.assetRepository.FindByIdAndDecrypt(assets[i].ID)
		if err != nil {
			msgChan <- fmt.Sprintf("资产「%v」Shell执行失败，查询数据异常「%v」", assets[i].Name, err.Error())
			return
		}

		var (
			username   = asset.Username
			password   = asset.Password
			privateKey = asset.PrivateKey
			passphrase = asset.Passphrase
			ip         = asset.IP
			port       = asset.Port
		)

		if asset.AccountType == "credential" {
			credential, err := r.jobService.credentialRepository.FindByIdAndDecrypt(asset.CredentialId)
			if err != nil {
				msgChan <- fmt.Sprintf("资产「%v」Shell执行失败，查询授权凭证数据异常「%v」", assets[i].Name, err.Error())
				return
			}

			if credential.Type == constants.Custom {
				username = credential.Username
				password = credential.Password
			} else {
				username = credential.Username
				privateKey = credential.PrivateKey
				passphrase = credential.Passphrase
			}
		}

		go func() {

			t1 := time.Now()
			result, err := ExecCommandBySSH(metadataShell.Shell, ip, port, username, password, privateKey, passphrase)
			elapsed := time.Since(t1)
			var msg string
			if err != nil {
				msg = fmt.Sprintf("资产「%v」Shell执行失败，返回值「%v」，耗时「%v」", asset.Name, err.Error(), elapsed)
				zlog.Info(msg)
			} else {
				msg = fmt.Sprintf("资产「%v」Shell执行成功，返回值「%v」，耗时「%v」", asset.Name, result, elapsed)
				zlog.Info(msg)
			}

			msgChan <- msg
		}()
	}

	var message = ""
	for i := 0; i < len(assets); i++ {
		message += <-msgChan + "\n"
	}

	_ = r.jobService.jobRepository.UpdateLastUpdatedById(r.ID)
	jobLog := models.JobLog{
		ID:      zos.GenUUID(),
		JobId:   r.ID,
		Message: message,
	}

	_ = r.jobService.jobLogRepository.Create(&jobLog)
}

func ExecCommandBySSH(cmd, ip string, port int, username, password, privateKey, passphrase string) (result string, err error) {
	sshClient, err := term.NewSshClient(ip, port, username, password, privateKey, passphrase)
	if err != nil {
		return "", err
	}

	session, err := sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	//执行远程命令
	combo, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}
	return string(combo), nil
}

func (r JobService) ExecJobById(id string) (err error) {
	job, err := r.jobRepository.FindById(id)
	if err != nil {
		return err
	}
	j, err := getJob(&job, &r)
	if err != nil {
		return err
	}
	j.Run()
	return nil
}

func (r JobService) LoadJobs() error {
	jobs, _ := r.jobRepository.FindByFunc(constants.FuncCheckAssetStatusJob)
	if len(jobs) == 0 {
		jobvm := models.Job{
			ID:     zos.GenUUID(),
			Name:   "虚拟机状态检测",
			Func:   constants.FuncCheckAssetStatusJob,
			Cron:   "@every 3m",
			Mode:   constants.JobModeAll,
			Status: constants.JobStatusRunning,
		}
		if err := r.jobRepository.Create(&jobvm); err != nil {
			return err
		}
		zlog.Debug("创建计划任务「%v」cron「%v」", jobvm.Name, jobvm.Cron)
		jobcluster := models.Job{
			ID:     zos.GenUUID(),
			Name:   "集群状态检测",
			Func:   constants.FuncCheckClusterStatusJob,
			Cron:   "@every 3m",
			Mode:   constants.JobModeAll,
			Status: constants.JobStatusRunning,
		}
		if err := r.jobRepository.Create(&jobcluster); err != nil {
			return err
		}
		zlog.Debug("创建计划任务「%v」cron「%v」", jobcluster.Name, jobcluster.Cron)
	} else {
		for i := range jobs {
			if jobs[i].Status == constants.JobStatusRunning {
				err := r.ChangeStatusById(jobs[i].ID, constants.JobStatusRunning)
				if err != nil {
					return err
				}
				zlog.Debug("启动计划任务「%v」cron「%v」", jobs[i].Name, jobs[i].Cron)
			}
		}
	}
	return nil
}

func (r JobService) Create(o *models.Job) (err error) {

	if o.Status == constants.JobStatusRunning {
		j, err := getJob(o, &r)
		if err != nil {
			return err
		}
		jobId, err := cronv1.Cron.AddJob(o.Cron, j)
		if err != nil {
			return err
		}
		o.CronJobId = int(jobId)
	}

	return r.jobRepository.Create(o)
}

func (r JobService) DeleteJobById(id string) error {
	job, err := r.jobRepository.FindById(id)
	if err != nil {
		return err
	}
	if job.Status == constants.JobStatusRunning {
		if err := r.ChangeStatusById(id, constants.JobStatusNotRunning); err != nil {
			return err
		}
	}
	return r.jobRepository.DeleteJobById(id)
}
