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

func (s *UserService) ProcessJoin(chatMemberUpdated *tgbotapi.ChatMemberUpdated) (*types.User, error) {
	err, _, log, _ := s.core.GetAll()

	tgUser := chatMemberUpdated.NewChatMember.User

	var dbUser *types.User
	if dbUser, err = s.userRepository.GetByTelegramID(tgUser.ID); err != nil {
		return nil, fmt.Errorf("failed to get user by telegram id: %w", err)
	}
	if dbUser != nil {
		if dbUser.InvitedByLinkID != nil {
			_, err = s.inviteLinkRepository.IncreaseCounterByID(*dbUser.InvitedByLinkID)
			if err != nil {
				return nil, fmt.Errorf("failed to IncreaseCounterByID for link [%d]: %v", *dbUser.InvitedByLinkID, err)
			}
		}

		userToUpdate := &types.User{
			TelegramID:   tgUser.ID,
			FirstName:    tgUser.FirstName,
			LastName:     tgUser.LastName,
			Username:     tgUser.UserName,
			MemberStatus: types.MemberStatus(chatMemberUpdated.NewChatMember.Status),
		}

		var updatedUser *types.User
		if updatedUser, err = s.userRepository.UpdateUserInfo(userToUpdate); err != nil {
			return nil, fmt.Errorf("failed to update user by telegram id: %w", err)
		}
		return updatedUser, nil
	}

	joinedUser := &types.User{
		TelegramID:   tgUser.ID,
		FirstName:    tgUser.FirstName,
		LastName:     tgUser.LastName,
		Username:     tgUser.UserName,
		MemberStatus: types.MemberStatus(chatMemberUpdated.NewChatMember.Status),
	}

	tgInviteLink := chatMemberUpdated.InviteLink
	if tgInviteLink != nil {
		if chatMemberUpdated.InviteLink.Name == "" {
			log.Warn("User [%s, %d] joined by noname invite link. URL: %s", joinedUser.FirstName, joinedUser.ID, chatMemberUpdated.InviteLink.InviteLink)
		} else {
			var inviteLink *types.InviteLink
			if inviteLink, err = s.inviteLinkRepository.GetByName(tgInviteLink.Name); err != nil {
				return nil, fmt.Errorf("failed to get invite link by name [%s]: %w", tgInviteLink.Name, err)
			} else if inviteLink == nil {
				log.Warn("User joined by unknown invite link. Name: %s, URL: %s", tgInviteLink.Name, tgInviteLink.InviteLink)
			} else {
				invitedByLinkID := inviteLink.ID
				joinedUser.InvitedByLinkID = &invitedByLinkID

				invitedByUserID := inviteLink.RequesterID
				joinedUser.InvitedByUserID = &invitedByUserID

				if _, err := s.inviteLinkRepository.IncreaseCounterByID(inviteLink.ID); err != nil {
					return nil, fmt.Errorf("failed to IncreaseCounterByID for invite link [%d, %s]: %v", inviteLink.ID, inviteLink.Name, err)
				}
			}
		}
	}

	var insertedUser *types.User
	if insertedUser, err = s.userRepository.Insert(joinedUser); err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}

	return insertedUser, nil
}

func (s *UserService) ProcessLeave(chatMemberUpdated *tgbotapi.ChatMemberUpdated) (*types.User, error) {
	err, _, log, _ := s.core.GetAll()

	tgUser := chatMemberUpdated.NewChatMember.User

	var dbUser *types.User
	if dbUser, err = s.userRepository.GetByTelegramID(tgUser.ID); err != nil {
		return nil, fmt.Errorf("failed to get user by telegram id: %w", err)
	}
	if dbUser != nil {
		if dbUser.InvitedByLinkID != nil {
			_, err = s.inviteLinkRepository.DecreaseCounterByID(*dbUser.InvitedByLinkID)
			if err != nil {
				return nil, fmt.Errorf("failed to DecreaseCounterByID for link [%d]: %v", *dbUser.InvitedByLinkID, err)
			}
		}

		userToUpdate := &types.User{
			TelegramID:   tgUser.ID,
			FirstName:    tgUser.FirstName,
			LastName:     tgUser.LastName,
			Username:     tgUser.UserName,
			MemberStatus: types.MemberStatus(chatMemberUpdated.NewChatMember.Status),
		}

		var updatedUser *types.User
		if updatedUser, err = s.userRepository.UpdateUserInfo(userToUpdate); err != nil {
			return nil, fmt.Errorf("failed to update user by telegram id: %w", err)
		}
		return updatedUser, nil
	}

	leftUser := &types.User{
		TelegramID:   tgUser.ID,
		Username:     tgUser.UserName,
		FirstName:    tgUser.FirstName,
		LastName:     tgUser.LastName,
		MemberStatus: types.MemberStatus(chatMemberUpdated.NewChatMember.Status),
	}

	var insertedUser *types.User
	insertedUser, err = s.userRepository.Insert(leftUser)
	if err != nil {
		return nil, fmt.Errorf("failed to Insert user: %w", err)
	}

	log.Info("Successfully processed leave for user @%s", insertedUser.Username)

	return insertedUser, nil
}

func (s *UserService) RequestMemberStatusByTelegramID(telegramID int64) (types.MemberStatus, error) {
	err, botAPI, _, _ := s.core.GetAll()

	var activeChannel *types.Channel
	if activeChannel, err = s.activeChannelRepository.Get(); err != nil {
		return "", fmt.Errorf("failed to get active channel: %w", err)
	}

	chatMember, err := botAPI.GetChatMember(
		tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: activeChannel.TelegramID,
				UserID: telegramID,
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("Failed to GetChatMember for user [%d] in channel [%d]: %v", telegramID, activeChannel.TelegramID, err)
	}

	return types.MemberStatus(chatMember.Status), nil
}

func (s *UserService) UpdateOrRegister(tgUser *tgbotapi.User) (*types.User, error) {
	if tgUser == nil {
		return nil, fmt.Errorf("tgUser cannot be nil")
	}

	err, _, _, _ := s.core.GetAll()

	userToInsertorUpdate := &types.User{
		TelegramID: tgUser.ID,
		FirstName:  tgUser.FirstName,
		LastName:   tgUser.LastName,
		Username:   tgUser.UserName,
	}
	userToInsertorUpdate.MemberStatus, err = s.RequestMemberStatusByTelegramID(tgUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to RequestMemberStatusFromTelegramID for user [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
	}

	var insertedUser *types.User
	if insertedUser, err = s.userRepository.InsertOrUpdateUserInfo(userToInsertorUpdate); err != nil {
		return nil, fmt.Errorf("Failed to Insert User [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
	}

	return insertedUser, nil
}

func (s *UserService) GetFromTGUser(tgUser *tgbotapi.User) (*types.User, error) {
	if tgUser == nil {
		return nil, fmt.Errorf("tgUser cannot be nil")
	}

	err, _, _, _ := s.core.GetAll()

	user, err := s.userRepository.GetByTelegramID(tgUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to GetByTelegramID user [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
	}
	return user, nil
}

/*
		func (s *UserService) UpdateFromTGUser(tgUser *tgbotapi.User) (*types.User, error) {
		if tgUser == nil {
			return nil, fmt.Errorf("tgUser cannot be nil")
		}

		err, _, log, _ := s.core.GetAll()
		if err != nil {
			return nil, err
			}

			var activeChannel *types.Channel
			if activeChannel, err = s.activeChannelRepository.Get(); err != nil {
			return nil, fmt.Errorf("failed to Get active channel: %w", err)
			} else if activeChannel == nil {
				return nil, fmt.Errorf("active channel is nil")
		}

		var userToUpdate *types.User
		userToUpdate, err = s.userRepository.GetByTelegramID(tgUser.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		} else if userToUpdate == nil {
			return nil, fmt.Errorf("couldn't find user: [%s, %d]", tgUser.FirstName, tgUser.ID)
		}

		userToUpdate.Username = tgUser.UserName
		userToUpdate.FirstName = tgUser.FirstName
		userToUpdate.LastName = tgUser.LastName
		userToUpdate.MemberStatus, err = s.RequestMemberStatusFromTelegramID(activeChannel.TelegramID, tgUser.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to RequestMemberStatusFromTelegramID for user [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
		}

		updatedUser, err := s.userRepository.UpdateBasedOnTelegramID(userToUpdate)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert user: %w", err)
		}

		log.Info("Updated new user: @%s (ID: %d)", updatedUser.Username, updatedUser.ID)
		return updatedUser, nil
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

*/
