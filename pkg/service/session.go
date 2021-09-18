package service

import (
	"next-terminal/constants"
	"next-terminal/models"
	"next-terminal/pkg/utils"
	"next-terminal/repository"
)

type SessionService struct {
	sessionRepository *repository.SessionRepository
}

func NewSessionService(sessionRepository *repository.SessionRepository) *SessionService {
	return &SessionService{sessionRepository: sessionRepository}
}

func (r SessionService) FixSessionState() error {
	sessions, err := r.sessionRepository.FindByStatus(constants.Connected)
	if err != nil {
		return err
	}

	if len(sessions) > 0 {
		for i := range sessions {
			session := models.Session{
				Status:           constants.Disconnected,
				DisconnectedTime: utils.NowJsonTime(),
			}

			_ = r.sessionRepository.UpdateById(&session, sessions[i].ID)
		}
	}
	return nil
}

func (r SessionService) EmptyPassword() error {
	return r.sessionRepository.EmptyPassword()
}
