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

func (s *InviteLinkService) CreateForRequester(requester *types.User) (*types.InviteLink, error) {
	err, botAPI, log, _ := s.core.GetAll()
	if err != nil {
		return nil, err
	}

	activeChannel, err := s.activeChannelRepository.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get active channel: %w", err)
	}

	if activeChannel == nil {
		return nil, fmt.Errorf("no active channel, ignoring")
	}

	var alreadyCreatedLink *types.InviteLink
	if alreadyCreatedLink, err = s.GetByRequester(requester); err != nil {
		return nil, fmt.Errorf("failed to get by requester: %w", err)
	}
	if alreadyCreatedLink != nil {
		return nil, fmt.Errorf("link already created")
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
	requester.InviteLinkID = &inviteLinkID
	s.userRepository.UpsertBasedOnTelegramID(requester)

	log.Info("Created invite link for channel %s: %s", activeChannel.Title, createdLink.Name)

	return inviteLink, nil
}

func (s *InviteLinkService) generateLinkName(requester *types.User) string {
	return fmt.Sprintf("%d", requester.TelegramID)
}

func (s *InviteLinkService) GetByRequester(requester *types.User) (*types.InviteLink, error) {
	err, _, _, _ := s.core.GetAll()

	if requester.InviteLinkID == nil {
		return nil, nil
	}

	var inviteLink *types.InviteLink
	if inviteLink, err = s.inviteLinkRepository.GetByID(*requester.InviteLinkID); err != nil {
		return nil, fmt.Errorf("failed to get invite link by id: %w", err)
	}

	return inviteLink, nil
}
