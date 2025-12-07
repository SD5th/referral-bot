package handlers

import (
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type myChatMemberHandler struct {
	core *core.Core

	tgUtilsService         interfaces.TGUtilsService
	activeChannelService   interfaces.ActiveChannelService
	userService            interfaces.UserService
	adminService           interfaces.AdminService
	inviteLinkService      interfaces.InviteLinkService
	channelActivityService interfaces.ChannelActivityService
}

func (h *myChatMemberHandler) Handle(myChatMemberUpdated *tgbotapi.ChatMemberUpdated) {
	log := h.core.GetLogger()

	log.Info("MyChatMember update: in chat %d, status: %s -> %s",
		myChatMemberUpdated.Chat.ID,
		myChatMemberUpdated.OldChatMember.Status,
		myChatMemberUpdated.NewChatMember.Status,
	)

	switch {
	case h.isJoinEvent(myChatMemberUpdated):
		h.handleJoin(myChatMemberUpdated)
	case h.isLeaveEvent(myChatMemberUpdated):
		h.handleLeave(myChatMemberUpdated)
	default:
		log.Info("Ignoring my chat member update - no relevant status change")
	}

}

func (h *myChatMemberHandler) isJoinEvent(chatMemberUpdated *tgbotapi.ChatMemberUpdated) bool {
	old := types.MemberStatus(chatMemberUpdated.OldChatMember.Status)
	new := types.MemberStatus(chatMemberUpdated.NewChatMember.Status)

	if old.NotInChannel() && new.InChannel() {
		return true
	}
	return false
}

func (h *myChatMemberHandler) isLeaveEvent(chatMemberUpdated *tgbotapi.ChatMemberUpdated) bool {
	old := types.MemberStatus(chatMemberUpdated.OldChatMember.Status)
	new := types.MemberStatus(chatMemberUpdated.NewChatMember.Status)

	if old.InChannel() && new.NotInChannel() {
		return true
	}
	return false
}

func (h *myChatMemberHandler) handleJoin(myChatMemberUpdated *tgbotapi.ChatMemberUpdated) {
	log := h.core.GetLogger()
	inviter := myChatMemberUpdated.From
	channel := myChatMemberUpdated.Chat
	log.Info("User [%s, %d] invited bot to channel [%s, %d]", inviter.FirstName, inviter.ID, channel.Title, channel.ID)
	/*
		isAdmin, err := h.adminService.IsAdmin(inviterTelegramID)
		if err != nil {
			log.Warn("Failed to check admin status: %v", err)
			log.Info("Leaving chat")
			h.leaveChannel(channelID)
			return
		}

		if !isAdmin {
			log.Info("Inviter is not admin, leaving chat")
			h.leaveChannel(channelID)
			return
		}

		activeChannel, err := h.activeChannelService.Get()
		if err != nil {
			log.Warn("Failed to get active channel: %v", err)
			log.Info("Leaving chat")
			h.leaveChannel(channelID)
			return
		}

		if activeChannel != nil {
			log.Info("Already have active channel, leaving new chat")
			h.leaveChannel(channelID)
			return
		}

		if _, err := h.activeChannelService.Register(channelID); err != nil {
			log.Warn("Failed to register new active channel: %v", err)
			h.leaveChannel(channelID)
			return
		}
	*/
}

func (h *myChatMemberHandler) handleLeave(myChatMemberUpdated *tgbotapi.ChatMemberUpdated) {
	log := h.core.GetLogger()
	kicker := myChatMemberUpdated.From
	channel := myChatMemberUpdated.Chat
	log.Info("User [%s, %d] kicked bot from channel [%s, %d]", kicker.FirstName, kicker.ID, channel.FirstName, channel.ID)
}

/*
	activeChannel, err := h.activeChannelService.Get()
	if err != nil {
		log.Error("Failed to get active channel: %v", err)
		return
	}

	if activeChannel == nil {
		log.Info("Have no active chat, ignoring")
		return
	}

	if activeChannel.TelegramID != myChatMemberUpdated.Chat.ID {
		log.Info("Kicked from non-active chat, ignoring")
		return
	}

	log.Warn("Kicked from active chat, you beter be joking boy")

	func (h *myChatMemberHandler) leaveChannel(channelID int64) error {
		err, botAPI, log, _ := h.core.GetAll()
	if err != nil {
		return err
	}

	_, err = botAPI.Request(tgbotapi.LeaveChatConfig{ChatID: channelID})
	if err != nil {
		return fmt.Errorf("failed to leave channel: %w", err)
	}

	log.Info("Bot left channel: %d", channelID)
	return nil
}

*/
