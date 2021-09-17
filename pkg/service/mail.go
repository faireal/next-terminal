package service

import (
	"github.com/ergoapi/zlog"
	"net/smtp"
	"next-terminal/repository"

	"github.com/jordan-wright/email"
	"next-terminal/pkg/constant"
)

type MailService struct {
	configsRepository *repository.ConfigsRepository
}

func NewMailService(configsRepository *repository.ConfigsRepository) *MailService {
	return &MailService{configsRepository: configsRepository}
}

func (r MailService) SendMail(to, subject, text string) {
	cfgsMap := r.configsRepository.FindAllMap()
	host := cfgsMap[constant.MailHost]
	port := cfgsMap[constant.MailPort]
	username := cfgsMap[constant.MailUsername]
	password := cfgsMap[constant.MailPassword]

	if host == "" || port == "" || username == "" || password == "" {
		zlog.Debug("邮箱信息不完整，跳过发送邮件。")
		return
	}

	e := email.NewEmail()
	e.From = "Next Terminal <" + username + ">"
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(text)
	err := e.Send(host+":"+port, smtp.PlainAuth("", username, password, host))
	if err != nil {
		zlog.Error("邮件发送失败: %v", err.Error())
	}
}
