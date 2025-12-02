package handlers

import (
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UpdateHandler struct {
	core *core.Core

	messageHandler      *messageHandler
	chatMemberHandler   *chatMemberHandler
	myChatMemberHandler *myChatMemberHandler

	tgUtilsService       interfaces.TGUtilsService
	activeChannelService interfaces.ActiveChannelService
	userService          interfaces.UserService
	adminService         interfaces.AdminService
	inviteLinkService    interfaces.InviteLinkService
}

func NewUpdateHandler(
	core *core.Core,
	activeChannelService interfaces.ActiveChannelService,
	userService interfaces.UserService,
	adminService interfaces.AdminService,
	tgUtilsService interfaces.TGUtilsService,
	inviteLinkService interfaces.InviteLinkService,

) (*UpdateHandler, error) {
	return &UpdateHandler{
		core: core,
		messageHandler: &messageHandler{
			core:                 core,
			activeChannelService: activeChannelService,
			tgUtilsService:       tgUtilsService,
			userService:          userService,
			adminService:         adminService,
			inviteLinkService:    inviteLinkService,
		},
		chatMemberHandler: &chatMemberHandler{
			core:                 core,
			activeChannelService: activeChannelService,
			userService:          userService,
			adminService:         adminService,
			inviteLinkService:    inviteLinkService,
		},
		myChatMemberHandler: &myChatMemberHandler{
			core:                 core,
			activeChannelService: activeChannelService,
			tgUtilsService:       tgUtilsService,
			userService:          userService,
			adminService:         adminService,
			inviteLinkService:    inviteLinkService,
		},
	}, nil
}

func (u *UpdateHandler) HandleUpdate(update tgbotapi.Update) error {
	defer func() {
		if r := recover(); r != nil {
			u.core.GetLogger().Error("Panic in HandleUpdate: %v", r)
		}
	}()

	switch {
	case update.Message != nil:
		u.messageHandler.Handle(update.Message)
	case update.MyChatMember != nil:
		u.myChatMemberHandler.Handle(update.MyChatMember)
	case update.ChatMember != nil:
		u.chatMemberHandler.Handle(update.ChatMember)
	}
	return nil
}
