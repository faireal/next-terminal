package service

import (
	"next-terminal/models"
	"next-terminal/repository"
	"strconv"
)

type NumService struct {
	numRepository *repository.NumRepository
}

func NewNumService(numRepository *repository.NumRepository) *NumService {
	return &NumService{numRepository: numRepository}
}

func (r NumService) InitNums() error {
	nums, err := r.numRepository.FindAll()
	if err != nil {
		return err
	}
	if len(nums) == 0 {
		for i := 0; i <= 30; i++ {
			if err := r.numRepository.Create(&models.Num{I: strconv.Itoa(i)}); err != nil {
				return err
			}
		}
	}
	return nil
}
