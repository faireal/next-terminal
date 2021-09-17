package service

import (
	"next-terminal/models"
	"next-terminal/pkg/utils"
	repository2 "next-terminal/repository"
)

type AssetService struct {
	assetRepository *repository2.AssetRepository
	userRepository  *repository2.UserRepository
}

func NewAssetService(assetRepository *repository2.AssetRepository, userRepository *repository2.UserRepository) *AssetService {
	return &AssetService{assetRepository: assetRepository, userRepository: userRepository}
}

func (r AssetService) Encrypt() error {
	items, err := r.assetRepository.FindAll()
	if err != nil {
		return err
	}
	for i := range items {
		item := items[i]
		if item.Encrypted {
			continue
		}
		if err := r.assetRepository.Encrypt(&item, utils.Encryption()); err != nil {
			return err
		}
		if err := r.assetRepository.UpdateById(&item, item.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r AssetService) InitDemoVM() error {
	u, err := r.userRepository.UserGet("username = ?", "admin")
	if err != nil {
		return err
	}
	debian := models.Asset{}
	debian.ID = utils.UUID()
	debian.Name = "debian"
	debian.Protocol = "ssh"
	debian.IP = "debian.ysicing.svc"
	debian.Port = 22
	debian.AccountType = "account_type"
	debian.Username = "root"
	debian.Password = "next-terminal"
	debian.Description = "默认演示debian"
	debian.Tags = "debian"
	debian.Owner = u.ID
	debian.Encrypted = true
	debian.Created = utils.NowJsonTime()

	m := map[string]interface{}{
		"ssh-mode": "guacd",
	}
	if err := r.assetRepository.InitAsset(&debian, m); err != nil {
		return err
	}
	centos := models.Asset{}
	centos.ID = utils.UUID()
	centos.Name = "debian"
	centos.Protocol = "ssh"
	centos.IP = "centos.ysicing.svc"
	centos.Port = 22
	centos.AccountType = "account_type"
	centos.Username = "next-terminal"
	centos.Password = "next-terminal"
	centos.Description = "默认演示centos"
	centos.Tags = "centos"
	centos.Owner = u.ID
	centos.Encrypted = true
	centos.Created = utils.NowJsonTime()
	return r.assetRepository.InitAsset(&centos, m)
}
