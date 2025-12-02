package services

import (
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserService struct {
	core *core.Core

	activeChannelRepository   interfaces.ActiveChannelRepository
	userRepository            interfaces.UserRepository
	inviteLinkRepository      interfaces.InviteLinkRepository
	channelActivityRepository interfaces.ChannelActivityRepository
}

func NewUserService(
	core *core.Core,
	activeChannelRepository interfaces.ActiveChannelRepository,
	userRepository interfaces.UserRepository,
	inviteLinkRepository interfaces.InviteLinkRepository,
	channelActivityRepository interfaces.ChannelActivityRepository,
) *UserService {
	return &UserService{
		core:                      core,
		activeChannelRepository:   activeChannelRepository,
		userRepository:            userRepository,
		inviteLinkRepository:      inviteLinkRepository,
		channelActivityRepository: channelActivityRepository,
	}
}

func (s *UserService) RegisterInDB(tgUser *tgbotapi.User) (*types.User, error) {
	if tgUser == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	err, botAPI, log, _ := s.core.GetAll()

	var activeChannel *types.Channel
	if activeChannel, err = s.activeChannelRepository.Get(); err != nil {
		return nil, fmt.Errorf("failed to get active channel: %w", err)
	} else if activeChannel == nil {
		return nil, fmt.Errorf("active channel is nil")
	}

	if dbUser, err := s.userRepository.GetByTelegramID(tgUser.ID); err != nil {
		return nil, fmt.Errorf("Failed to GetByTelegramID User [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
	} else if dbUser != nil {
		return nil, fmt.Errorf("user is already in db: [%s, %d]", dbUser.FirstName, dbUser.ID)
	}

	userToRegister := new(types.User)
	userToRegister.TelegramID = tgUser.ID
	userToRegister.FirstName = tgUser.FirstName
	userToRegister.LastName = tgUser.LastName
	userToRegister.Username = tgUser.UserName

	chatMember, err := botAPI.GetChatMember(
		tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: activeChannel.TelegramID,
				UserID: tgUser.ID,
			},
		},
	)
	if err != nil {
		log.Error("Failed to get ChatMember for user [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
		userToRegister.MemberStatus = types.MemberStatusLeft
	} else {
		userToRegister.MemberStatus = types.MemberStatus(chatMember.Status)
	}

	var registeredUser *types.User
	if registeredUser, err = s.userRepository.Insert(userToRegister); err != nil {
		return nil, fmt.Errorf("Failed to Insert User [%s, %d]: %v", userToRegister.FirstName, userToRegister.ID, err)
	}

	return registeredUser, nil
}

func (s *UserService) ProcessJoin(chatMemberUpdated *tgbotapi.ChatMemberUpdated) (*types.User, error) {
	err, _, log, _ := s.core.GetAll()

	var activeChannel *types.Channel
	if activeChannel, err = s.activeChannelRepository.Get(); err != nil {
		return nil, fmt.Errorf("failed to get active channel: %w", err)
	} else if activeChannel.TelegramID != chatMemberUpdated.Chat.ID {
		return nil, fmt.Errorf("user joined non-active channel, please leave non-active channels")
	}

	tgUser := chatMemberUpdated.NewChatMember.User

	var joinedUser *types.User
	if joinedUser, err = s.userRepository.GetByTelegramID(tgUser.ID); err != nil {
		return nil, fmt.Errorf("failed to get user by telegram id: %w", err)
	}
	if joinedUser == nil {
		joinedUser = new(types.User)
		joinedUser.TelegramID = tgUser.ID
	}

	joinedUser.Username = tgUser.UserName
	joinedUser.FirstName = tgUser.FirstName
	joinedUser.LastName = tgUser.LastName
	joinedUser.MemberStatus = types.MemberStatus(chatMemberUpdated.NewChatMember.Status)

	if chatMemberUpdated.InviteLink != nil {
		if chatMemberUpdated.InviteLink.Name == "" {
			log.Warn("User joined by noname invite link. URL: %s", chatMemberUpdated.InviteLink.InviteLink)
		} else {
			var inviteLink *types.InviteLink
			if inviteLink, err = s.inviteLinkRepository.GetByName(chatMemberUpdated.InviteLink.Name); err != nil {
				return nil, fmt.Errorf("failed to get invite link by name: %w", err)
			}
			if inviteLink == nil {
				log.Warn("User joined by unknown invite link. Name: %s, URL: %s", chatMemberUpdated.InviteLink.Name, chatMemberUpdated.InviteLink.InviteLink)
			} else {
				invitedByLinkID := inviteLink.ID
				joinedUser.InvitedByLinkID = &invitedByLinkID

				invitedByUserID := inviteLink.RequesterID
				joinedUser.InvitedByUserID = &invitedByUserID

				inviteLink.UniqueJoins += 1
				if _, err := s.inviteLinkRepository.UpdateByID(inviteLink); err != nil {
					log.Warn("failed to update invite link: %v", err)
				}
			}
		}
	}

	var upsertedUser *types.User
	if upsertedUser, err = s.userRepository.UpsertBasedOnTelegramID(joinedUser); err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}

	activity := &types.ChannelActivity{
		UserID:    upsertedUser.ID,
		OldStatus: types.MemberStatus(chatMemberUpdated.OldChatMember.Status),
		NewStatus: types.MemberStatus(chatMemberUpdated.NewChatMember.Status),
	}
	if upsertedUser.InvitedByUserID != nil {
		invitedByUserID := *upsertedUser.InvitedByUserID
		activity.InvitedByUserID = &invitedByUserID
	}
	if upsertedUser.InvitedByLinkID != nil {
		invitedByLinkID := *upsertedUser.InvitedByLinkID
		activity.InvitedByLinkID = &invitedByLinkID
	}

	if _, err := s.channelActivityRepository.Insert(activity); err != nil {
		log.Warn("Failed to add channel activity: %v", err)
	}

	log.Info("Successfully processed join for user @%s", upsertedUser.Username)

	return upsertedUser, nil
}

func (s *UserService) ProcessLeave(chatMemberUpdated *tgbotapi.ChatMemberUpdated) (*types.User, error) {
	err, _, log, _ := s.core.GetAll()
	if err != nil {
		return nil, err
	}

	var activeChannel *types.Channel
	if activeChannel, err = s.activeChannelRepository.Get(); err != nil {
		return nil, fmt.Errorf("failed to get active channel: %w", err)
	} else if activeChannel.TelegramID != chatMemberUpdated.Chat.ID {
		return nil, fmt.Errorf("user left non-active channel, please leave non-active channels")
	}

	tgUser := chatMemberUpdated.NewChatMember.User

	var leftUser *types.User
	if leftUser, err = s.userRepository.GetByTelegramID(tgUser.ID); err != nil {
		return nil, fmt.Errorf("failed to get user by telegram id: %w", err)
	}
	if leftUser == nil {
		leftUser = new(types.User)
		leftUser.TelegramID = tgUser.ID
	}

	leftUser.Username = tgUser.UserName
	leftUser.FirstName = tgUser.FirstName
	leftUser.LastName = tgUser.LastName
	leftUser.MemberStatus = types.MemberStatus(chatMemberUpdated.NewChatMember.Status)
	leftUser.InvitedByUserID = nil

	if leftUser.InvitedByLinkID != nil {
		var inviteLink *types.InviteLink
		if inviteLink, err = s.inviteLinkRepository.GetByID(*leftUser.InvitedByLinkID); err != nil {
			return nil, fmt.Errorf("failed to get invite link by id: %w", err)
		} else if inviteLink == nil {
			return nil, fmt.Errorf("failed to get invite link by id")
		}

		inviteLink.UniqueJoins -= 1
		if _, err := s.inviteLinkRepository.UpdateByID(inviteLink); err != nil {
			return nil, fmt.Errorf("failed to upsert invite link: %w", err)
		}
	}

	var upsertedUser *types.User
	upsertedUser, err = s.userRepository.UpsertBasedOnTelegramID(leftUser)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}

	activity := &types.ChannelActivity{
		UserID:    upsertedUser.ID,
		OldStatus: types.MemberStatus(chatMemberUpdated.OldChatMember.Status),
		NewStatus: types.MemberStatus(chatMemberUpdated.NewChatMember.Status),
	}
	if upsertedUser.InvitedByUserID != nil {
		invitedByUserID := *upsertedUser.InvitedByUserID
		activity.InvitedByUserID = &invitedByUserID
	}
	if upsertedUser.InvitedByLinkID != nil {
		invitedByLinkID := *upsertedUser.InvitedByLinkID
		activity.InvitedByLinkID = &invitedByLinkID
	}

	if _, err := s.channelActivityRepository.Insert(activity); err != nil {
		log.Warn("Failed to insert channel activity: %v", err)
	}

	log.Info("Successfully processed leave for user @%s", upsertedUser.Username)

	return upsertedUser, nil
}

func (s *UserService) CanCreateReferralLink(telegramID int64) (bool, error) {
	activeChannel, err := s.activeChannelRepository.Get()
	if err != nil {
		return false, fmt.Errorf("failed to get active channel: %w", err)
	}
	if activeChannel == nil {
		return false, fmt.Errorf("no active channel")
	}

	user, err := s.userRepository.GetByTelegramID(telegramID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil || !user.IsMember() {
		return false, nil
	}

	return true, nil
}

func (s *UserService) IsAdmin(telegramID int64) (bool, error) {
	user, err := s.userRepository.GetByTelegramID(telegramID)
	if err != nil {
		return false, fmt.Errorf("failed to get user by telegram id: %w", err)
	}
	if user == nil {
		return false, nil
	}
	if user.IsAdmin() {
		return true, nil
	}
	return false, nil
}

func (s *UserService) GetOrUpdateFromMessage(message *tgbotapi.Message) (*types.User, error) {
	if message.From == nil {
		return nil, fmt.Errorf("message has no user data")
	}

	err, _, log, _ := s.core.GetAll()
	if err != nil {
		return nil, err
	}

	tgUser := message.From

	dbUser, err := s.userRepository.GetByTelegramID(tgUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if dbUser == nil {
		return nil, nil
	}

	dbUser.Username = tgUser.UserName
	dbUser.FirstName = tgUser.FirstName
	dbUser.LastName = tgUser.LastName

	upsertedUser, err := s.userRepository.UpsertBasedOnTelegramID(dbUser)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}

	log.Info("Updated new user: @%s (ID: %d)", upsertedUser.Username, upsertedUser.ID)
	return upsertedUser, nil
}
