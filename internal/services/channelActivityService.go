package services

import (
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChannelActivityService struct {
	core *core.Core

	activeChannelRepository   interfaces.ActiveChannelRepository
	userRepository            interfaces.UserRepository
	adminRepository           interfaces.AdminRepository
	inviteLinkRepository      interfaces.InviteLinkRepository
	channelActivityRepository interfaces.ChannelActivityRepository
}

func NewChannelActivityService(
	core *core.Core,
	activeChannelRepository interfaces.ActiveChannelRepository,
	userRepository interfaces.UserRepository,
	adminRepository interfaces.AdminRepository,
	inviteLinkRepository interfaces.InviteLinkRepository,
	channelActivityRepository interfaces.ChannelActivityRepository,
) *ChannelActivityService {
	return &ChannelActivityService{
		core:                      core,
		activeChannelRepository:   activeChannelRepository,
		userRepository:            userRepository,
		adminRepository:           adminRepository,
		inviteLinkRepository:      inviteLinkRepository,
		channelActivityRepository: channelActivityRepository,
	}
}

func (s *ChannelActivityService) AddFromUpdate(chatMemberUpdated *tgbotapi.ChatMemberUpdated) (*types.ChannelActivity, error) {
	err, _, _, _ := s.core.GetAll()
	activeChannel, err := s.activeChannelRepository.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get active channel: %w", err)
	}
	if activeChannel == nil {
		return nil, fmt.Errorf("no active channel")
	}

	tgUser := chatMemberUpdated.NewChatMember.User
	activity := &types.ChannelActivity{
		ChannelTelegramID: activeChannel.TelegramID,

		UserTelegramID: tgUser.ID,
		UserFirstName:  tgUser.FirstName,
		UserLastName:   tgUser.LastName,
		UserUsername:   tgUser.UserName,

		OldStatus: types.MemberStatus(chatMemberUpdated.OldChatMember.Status),
		NewStatus: types.MemberStatus(chatMemberUpdated.NewChatMember.Status),
	}

	tgInviteLink := chatMemberUpdated.InviteLink
	if chatMemberUpdated.InviteLink != nil {
		activity.InviteLinkName = tgInviteLink.Name
		activity.InviteLinkURL = tgInviteLink.InviteLink

		dbInviteLink, err := s.inviteLinkRepository.GetByName(tgInviteLink.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get invite link by name: %w", err)
		}
		if dbInviteLink != nil {
			dbInviter, err := s.userRepository.GetByID(dbInviteLink.RequesterID)
			if err != nil {
				return nil, fmt.Errorf("failed to get inviter by id: %w", err)
			}
			if dbInviter != nil {
				InviterTelegramID := dbInviter.TelegramID
				activity.InviterTelegramID = &InviterTelegramID
				activity.InviterFirstName = dbInviter.FirstName
				activity.InviterLastName = dbInviter.LastName
				activity.InviterUsername = dbInviter.Username
			}
		}
	}
	var insertedActivity *types.ChannelActivity
	if insertedActivity, err = s.channelActivityRepository.Insert(activity); err != nil {
		return nil, fmt.Errorf("failed to insert channel activity: %v", err)
	}

	return insertedActivity, nil
}
