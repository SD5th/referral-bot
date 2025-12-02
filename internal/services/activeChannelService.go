package services

import (
	"encoding/json"
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ActiveChannelService struct {
	core *core.Core

	activeChannelRepository   interfaces.ActiveChannelRepository
	userRepository            interfaces.UserRepository
	inviteLinkRepository      interfaces.InviteLinkRepository
	channelActivityRepository interfaces.ChannelActivityRepository
}

func NewActiveChannelService(
	core *core.Core,
	activeChannelRepository interfaces.ActiveChannelRepository,
	userRepository interfaces.UserRepository,
	inviteLinkRepository interfaces.InviteLinkRepository,
	channelActivityRepository interfaces.ChannelActivityRepository,
) *ActiveChannelService {
	return &ActiveChannelService{
		core:                      core,
		activeChannelRepository:   activeChannelRepository,
		userRepository:            userRepository,
		inviteLinkRepository:      inviteLinkRepository,
		channelActivityRepository: channelActivityRepository,
	}
}

func (s *ActiveChannelService) Register(telegramID int64) (*types.Channel, error) {
	err, botAPI, log, _ := s.core.GetAll()
	if err != nil {
		return nil, err
	}
	log.Info("Attempting to register channel %d as active", telegramID)

	activeChannel, err := s.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get active channel: %w", err)
	}

	if activeChannel != nil {
		if activeChannel.TelegramID == telegramID {
			return activeChannel, nil // АДЕКВАТНО РАССМОТРЕТЬ СЛУЧАЙ
		}
		return nil, fmt.Errorf("already have an active channel [%d], ignoring call to register [%d]", activeChannel.TelegramID, telegramID)
	}

	if err := s.validateBotPermissionsInChannel(telegramID); err != nil {
		return nil, fmt.Errorf("permission validation failed: %w", err)
	}

	chat, err := botAPI.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{ChatID: telegramID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get channel info: %w", err)
	}

	inviteConfig := tgbotapi.CreateChatInviteLinkConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chat.ID,
		},
		Name:               "Main",
		MemberLimit:        0,
		ExpireDate:         0,
		CreatesJoinRequest: false,
	}

	resp, err := botAPI.Request(inviteConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create invite link: %w", err)
	}

	var chatInviteLink *tgbotapi.ChatInviteLink
	if err := json.Unmarshal(resp.Result, &chatInviteLink); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	newActiveChannel := new(types.Channel)
	newActiveChannel.TelegramID = chat.ID
	newActiveChannel.Title = chat.Title
	newActiveChannel.Type = types.ChatType(chat.Type)
	newActiveChannel.Username = chat.UserName
	newActiveChannel.InviteLink = chatInviteLink.InviteLink

	var dbActiveChannel *types.Channel
	if dbActiveChannel, err = s.activeChannelRepository.Set(newActiveChannel); err != nil {
		return nil, fmt.Errorf("failed to Set new active channel: %w", err)
	}

	log.Info("Channel info: %s (@%s)", chat.Title, chat.UserName)
	return dbActiveChannel, nil
}

func (s *ActiveChannelService) validateBotPermissionsInChannel(channelTelegramID int64) error {
	err, botAPI, _, _ := s.core.GetAll()
	if err != nil {
		return err
	}

	member, err := botAPI.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: channelTelegramID,
			UserID: botAPI.Self.ID,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to get chat member: %w", err)
	}

	if !member.IsAdministrator() {
		return fmt.Errorf("bot is not administrator in channel")
	}

	if member.CanInviteUsers == false {
		return fmt.Errorf("bot cannot invite users in channel %d", channelTelegramID)
	}

	return nil
}

func (s *ActiveChannelService) IsActive() bool {
	err, _, log, _ := s.core.GetAll()

	activeChannel, err := s.Get()
	if err == nil && activeChannel != nil {
		return true
	}
	if err != nil {
		log.Error("Failed to GetActiveChannel: %v", err)
	}
	return false
}

func (s *ActiveChannelService) Get() (*types.Channel, error) {
	channel, err := s.activeChannelRepository.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get active channel: %w", err)
	}
	return channel, nil
}

func (s *ActiveChannelService) Leave() error {
	err, botAPI, log, _ := s.core.GetAll()
	if err != nil {
		return err
	}

	activeChannel, err := s.Get()
	if err != nil {
		return fmt.Errorf("failed to get active channel: %w", err)
	}

	if activeChannel == nil {
		return fmt.Errorf("no active channel, ignoring")
	}

	_, err = botAPI.Request(tgbotapi.LeaveChatConfig{ChatID: activeChannel.TelegramID})
	if err != nil {
		return fmt.Errorf("failed to leave channel: %w", err)
	}

	// ОЧИСТИТЬ БАЗУ ДАННЫХ

	log.Info("Bot left active channel: %d", activeChannel.TelegramID)
	return nil
}
