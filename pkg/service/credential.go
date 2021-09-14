package service

import (
	"next-terminal/server/repository"
	"next-terminal/server/utils"
)

type CredentialService struct {
	credentialRepository *repository.CredentialRepository
}

func NewCredentialService(credentialRepository *repository.CredentialRepository) *CredentialService {
	return &CredentialService{credentialRepository: credentialRepository}
}

func (r CredentialService) Encrypt() error {
	items, err := r.credentialRepository.FindAll()
	if err != nil {
		return err
	}
	for i := range items {
		item := items[i]
		if item.Encrypted {
			continue
		}
		if err := r.credentialRepository.Encrypt(&item, utils.Encryption()); err != nil {
			return err
		}
		if err := r.credentialRepository.UpdateById(&item, item.ID); err != nil {
			return err
		}
	}
	return nil
}
