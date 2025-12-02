package handlers

import (
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type chatMemberHandler struct {
	core *core.Core

	tgUtilsServiceService  interfaces.TGUtilsService
	activeChannelService   interfaces.ActiveChannelService
	userService            interfaces.UserService
	adminService           interfaces.AdminService
	inviteLinkService      interfaces.InviteLinkService
	channelActivityService interfaces.ChannelActivityService
}

func (h *chatMemberHandler) Handle(chatMemberUpdated *tgbotapi.ChatMemberUpdated) {
	log := h.core.GetLogger()

	if h.isUserBot(chatMemberUpdated) {
		return
	}

	activeChannel, err := h.activeChannelService.Get()
	if err != nil {
		log.Error("Failed to get active channel: %v", err)
		return
	}

	if activeChannel == nil || activeChannel.TelegramID != chatMemberUpdated.Chat.ID {
		log.Warn("Ignoring event from non-active channel: %d. You better leave it..", chatMemberUpdated.Chat.ID)
		return
	}

	var channelActivity *types.ChannelActivity
	if channelActivity, err = h.channelActivityService.AddFromUpdate(chatMemberUpdated); err != nil {
		log.Error("Failed to AddFromUpdate channel activity: %v", err)
	}

	log.Info("ChatMember update: %s, status: %s -> %s",
		channelActivity.UserFirstName,
		channelActivity.OldStatus,
		channelActivity.NewStatus,
	)

	switch {
	case h.isJoinEvent(chatMemberUpdated):
		h.handleJoin(chatMemberUpdated)
	case h.isLeaveEvent(chatMemberUpdated):
		h.handleLeave(chatMemberUpdated)
	default:
		log.Info("Ignoring chat member update - no relevant status change")
	}
}

func (h *chatMemberHandler) isUserBot(chatMemberUpdated *tgbotapi.ChatMemberUpdated) bool {
	return chatMemberUpdated.NewChatMember.User.IsBot || chatMemberUpdated.OldChatMember.User.IsBot
}

func (h *chatMemberHandler) isJoinEvent(chatMemberUpdated *tgbotapi.ChatMemberUpdated) bool {
	old := types.MemberStatus(chatMemberUpdated.OldChatMember.Status)
	new := types.MemberStatus(chatMemberUpdated.NewChatMember.Status)

	if (old == types.MemberStatusKicked) || (old == types.MemberStatusLeft) {
		if (new == types.MemberStatusAdministrator) || (new == types.MemberStatusMember) || (new == types.MemberStatusCreator) || (new == types.MemberStatusRestricted) {
			return true
		}
	}
	return false
}

func (h *chatMemberHandler) isLeaveEvent(chatMemberUpdated *tgbotapi.ChatMemberUpdated) bool {
	old := types.MemberStatus(chatMemberUpdated.OldChatMember.Status)
	new := types.MemberStatus(chatMemberUpdated.NewChatMember.Status)

	if (old == types.MemberStatusAdministrator) || (old == types.MemberStatusMember) || (old == types.MemberStatusCreator) || (old == types.MemberStatusRestricted) {
		if (new == types.MemberStatusKicked) || (new == types.MemberStatusLeft) {
			return true
		}
	}
	return false
}

func (h *chatMemberHandler) handleJoin(update *tgbotapi.ChatMemberUpdated) {
	log := h.core.GetLogger()

	user, err := h.userService.ProcessJoin(update)
	if err != nil {
		log.Error("Failed to process user join: %v", err)
		return
	}

	if user != nil {
		log.Info("Successfully processed join for user @%s (ID: %d)",
			update.NewChatMember.User.UserName, user.TelegramID)
	} else {
		log.Warn("ProcessJoin returned nil user for @%s",
			update.NewChatMember.User.UserName)
	}
}

func (h *chatMemberHandler) handleLeave(update *tgbotapi.ChatMemberUpdated) {
	log := h.core.GetLogger()

	user, err := h.userService.ProcessLeave(update)
	if err != nil {
		log.Error("Failed to process user leave: %v", err)
		return
	}
	if user != nil {
		log.Info("Successfully processed leave for user @%s (ID: %d)",
			update.NewChatMember.User.UserName, user.TelegramID)
	} else {
		log.Warn("ProcessLeave returned nil user for @%s",
			update.NewChatMember.User.UserName)
	}
}
