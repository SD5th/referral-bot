package services

import (
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminService struct {
	core *core.Core

	activeChannelRepository   interfaces.ActiveChannelRepository
	userRepository            interfaces.UserRepository
	adminRepository           interfaces.AdminRepository
	inviteLinkRepository      interfaces.InviteLinkRepository
	channelActivityRepository interfaces.ChannelActivityRepository
}

func NewAdminService(
	core *core.Core,
	activeChannelRepository interfaces.ActiveChannelRepository,
	userRepository interfaces.UserRepository,
	adminRepository interfaces.AdminRepository,
	inviteLinkRepository interfaces.InviteLinkRepository,
	channelActivityRepository interfaces.ChannelActivityRepository,
) *AdminService {
	return &AdminService{
		core:                      core,
		activeChannelRepository:   activeChannelRepository,
		userRepository:            userRepository,
		adminRepository:           adminRepository,
		inviteLinkRepository:      inviteLinkRepository,
		channelActivityRepository: channelActivityRepository,
	}
}

func (s *AdminService) Register(tgUser *tgbotapi.User) (*types.Admin, error) {
	err, _, log, _ := s.core.GetAll()
	if err != nil {
		return nil, err
	}

	var dbAdmin *types.Admin
	if dbAdmin, err = s.adminRepository.GetByTelegramID(tgUser.ID); err != nil {
		return nil, fmt.Errorf("failed to get admin by telegram id: %w", err)
	}
	if dbAdmin != nil {
		return nil, fmt.Errorf("user with Telegram ID: %d is already registered as admin", tgUser.ID)
	}

	newAdmin := &types.Admin{
		TelegramID: tgUser.ID,
		FirstName:  tgUser.FirstName,
		LastName:   tgUser.LastName,
		Username:   tgUser.UserName,
	}

	var insertedAdmin *types.Admin
	if insertedAdmin, err = s.adminRepository.Insert(newAdmin); err != nil {
		log.Warn("Failed to insert new admin: %v", err)
	}

	log.Info("Successfully registered admin @%s", insertedAdmin.Username)

	return insertedAdmin, nil
}
func (s *AdminService) IsAdmin(telegramID int64) (bool, error) {
	err, _, _, _ := s.core.GetAll()
	if err != nil {
		return false, err
	}

	var dbAdmin *types.Admin
	if dbAdmin, err = s.adminRepository.GetByTelegramID(telegramID); err != nil {
		return false, fmt.Errorf("failed to get admin by telegram id: %w", err)
	}
	if dbAdmin != nil {
		return true, nil
	}
	return false, nil
}
