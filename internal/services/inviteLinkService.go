package services

import (
	"encoding/json"
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type InviteLinkService struct {
	core *core.Core

	activeChannelRepository   interfaces.ActiveChannelRepository
	userRepository            interfaces.UserRepository
	inviteLinkRepository      interfaces.InviteLinkRepository
	channelActivityRepository interfaces.ChannelActivityRepository
}

func NewInviteLinkService(
	core *core.Core,
	activeChannelRepository interfaces.ActiveChannelRepository,
	userRepository interfaces.UserRepository,
	inviteLinkRepository interfaces.InviteLinkRepository,
	channelActivityRepository interfaces.ChannelActivityRepository,
) *InviteLinkService {
	return &InviteLinkService{
		core:                      core,
		activeChannelRepository:   activeChannelRepository,
		userRepository:            userRepository,
		inviteLinkRepository:      inviteLinkRepository,
		channelActivityRepository: channelActivityRepository,
	}
}

func (s *InviteLinkService) GetByRequesterTelegramID(telegramID int64) (*types.InviteLink, error) {
	err, _, _, _ := s.core.GetAll()

	var requester *types.User
	requester, err = s.userRepository.GetByTelegramID(telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to GetByTelegramID user: %w", err)
	} else if requester == nil {
		return nil, fmt.Errorf("there is no user with telegramID: %d", telegramID)
	}

	if requester.InviteLinkID == nil {
		return nil, nil
	}

	var createdLink *types.InviteLink
	if createdLink, err = s.inviteLinkRepository.GetByID(*requester.InviteLinkID); err != nil {
		return nil, fmt.Errorf("failed to get by requester: %w", err)
	} else if createdLink == nil {
		return nil, fmt.Errorf("cannot find link with id [%d]", *requester.InviteLinkID)
	}

	return createdLink, nil
}

func (s *InviteLinkService) GetOrCreateByRequesterTelegramID(telegramID int64) (*types.InviteLink, error) {
	err, botAPI, log, _ := s.core.GetAll()

	var activeChannel *types.Channel
	if activeChannel, err = s.activeChannelRepository.Get(); err != nil {
		return nil, fmt.Errorf("failed to get active channel: %w", err)
	} else if activeChannel == nil {
		return nil, fmt.Errorf("no active channel, ignoring")
	}

	var requester *types.User
	requester, err = s.userRepository.GetByTelegramID(telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to GetByTelegramID user: %w", err)
	} else if requester == nil {
		return nil, fmt.Errorf("there is no user with telegramID: %d", telegramID)
	}

	if requester.InviteLinkID != nil {
		var alreadyCreatedLink *types.InviteLink
		if alreadyCreatedLink, err = s.inviteLinkRepository.GetByID(*requester.InviteLinkID); err != nil {
			return nil, fmt.Errorf("failed to get by requester: %w", err)
		}
		if alreadyCreatedLink == nil {
			return nil, fmt.Errorf("alreadyCreatedLink is nil, but requester has not-nil InviteLinkID [%d]", *requester.InviteLinkID)
		}
	}

	linkName := s.generateLinkName(requester)
	inviteConfig := tgbotapi.CreateChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: activeChannel.TelegramID,
		},
		Name:               linkName,
		MemberLimit:        0,
		ExpireDate:         0,
		CreatesJoinRequest: false,
	}

	resp, err := botAPI.Request(inviteConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create invite link: %w", err)
	}

	var result *tgbotapi.ChatInviteLink
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	inviteLink := &types.InviteLink{
		RequesterID: requester.ID,
		URL:         result.InviteLink,
		Name:        result.Name,
		UniqueJoins: 0,
	}

	var createdLink *types.InviteLink
	if createdLink, err = s.inviteLinkRepository.Insert(inviteLink); err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	inviteLinkID := createdLink.ID
	userToUpdate := &types.User{
		TelegramID:   requester.TelegramID,
		InviteLinkID: &inviteLinkID,
	}

	if _, err = s.userRepository.UpdateLinkInfo(userToUpdate); err != nil {
		return nil, fmt.Errorf("failed to UpdateLinkInfo for user [%d]: %w", requester.TelegramID, err)
	}

	log.Info("Created invite link for channel %s: %s", activeChannel.Title, createdLink.Name)

	return inviteLink, nil
}

func (s *InviteLinkService) generateLinkName(requester *types.User) string {
	return fmt.Sprintf("%d", requester.TelegramID)
}
