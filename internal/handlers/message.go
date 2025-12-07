package handlers

import (
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"
	"referral-bot/internal/types"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type messageHandler struct {
	core *core.Core

	tgUtilsService         interfaces.TGUtilsService
	adminService           interfaces.AdminService
	activeChannelService   interfaces.ActiveChannelService
	userService            interfaces.UserService
	inviteLinkService      interfaces.InviteLinkService
	channelActivityService interfaces.ChannelActivityService
}

func (h *messageHandler) Handle(message *tgbotapi.Message) {
	log := h.core.GetLogger()
	if !h.isMessageFromUser(message) {
		log.Info("Ignoring non-user message:\n%s", message.Text)
		return
	}

	log.Info(""+
		"Processing message from: [%s]\n"+
		"%s",
		message.From.UserName, message.Text,
	)

	if message.IsCommand() {
		h.handleCommand(message)
		return
	}
	h.handleTextMessage(message)
}

func (h *messageHandler) isMessageFromUser(message *tgbotapi.Message) bool {
	if message.From == nil {
		return false
	}
	if message.From.IsBot {
		return false
	}
	return true
}

func (h *messageHandler) handleCommand(message *tgbotapi.Message) {
	log := h.core.GetLogger()
	command := message.Command()

	// User commands
	switch command {
	case "start":
		h.handleStartCommand(message)
		return
	case "referral":
		h.handleReferralCommand(message)
		return
	case "count":
		h.handleCountCommand(message)
		return
	case "help":
		h.handleHelpCommand(message)
		return

	}

	// Admin login command
	if command == "logadmin" {
		h.handleLogAdmin(message)
		return
	}

	isAdmin, err := h.adminService.IsAdmin(message.From.ID)
	if err != nil {
		log.Error("Failed to check admin status: %v", err)
		h.handleUnknownCommand(message)
		return
	}

	// Admin commands
	if isAdmin {
		switch command {
		case "regact":
			h.handleRegisterAsActive(message)
			return
		}
	}

	// Default
	h.handleUnknownCommand(message)
}

func (h *messageHandler) handleStartCommand(message *tgbotapi.Message) {
	_, _, log, _ := h.core.GetAll()

	if !h.activeChannelService.IsActive() {
		log.Warn("Start command aborted: No active channel.")
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	text := "" +
		"Привет! Я бот для розыгрыша.\n" +
		"Используй /help списка команд."

	h.tgUtilsService.SendMessage(message.Chat.ID, text)

}

func (h *messageHandler) handleReferralCommand(message *tgbotapi.Message) {
	err, _, log, _ := h.core.GetAll()

	// Getting and verifying active channel
	var activeChannel *types.Channel
	if activeChannel, err = h.activeChannelService.Get(); err != nil {
		log.Error("Referral command aborted: Failed to Get Active Channel: %v", err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	} else if activeChannel == nil {
		log.Warn("Referral command aborted: Active Channel is nil.")
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	log.Info("1")

	tgUser := message.From

	// Checking if user is a member
	var userMemberStatus types.MemberStatus
	userMemberStatus, err = h.userService.RequestMemberStatusByTelegramID(tgUser.ID)
	if err != nil {
		log.Error("Referral command aborted: Failed to RequestMemberStatusByTelegramID: %v", err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}
	if userMemberStatus.NotInChannel() {
		text := "" +
			"Не вижу тебя в подписчиках канала «Медиа Снабжение»!\n\n" +
			"Давай исправим, присоединяйся и к каналу, и к розыгрышу.\n"
		if activeChannel.InviteLink != "" {
			text += activeChannel.InviteLink
		} else {
			log.Warn("No available InviteLink in active channel")
		}

		h.tgUtilsService.SendMessage(message.Chat.ID, text)
		return
	}
	log.Info("2")

	// Adding user to DB, if he's not
	_, err = h.userService.UpdateOrRegister(tgUser)
	if err != nil {
		log.Error("Failed to UpdateOrRegister User [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}
	/*
		if user == nil {
			user, err = h.userService.AddFromTGUser(tgUser)
			if err != nil {
				log.Error("Failed to AddFromTGUser User [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
				h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
				return
			}
		}
	*/
	log.Info("3")

	// Getting or creating Invite Link
	var inviteLink *types.InviteLink
	inviteLink, err = h.inviteLinkService.GetOrCreateByRequesterTelegramID(tgUser.ID)
	if err != nil {
		log.Error("Failed to GetOrCreateByRequesterTelegramID link for user [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}
	log.Info("4")
	log.Info("%v", inviteLink)

	// Sending Link
	text := "" +
		"Пересылай эту ссылку коллегам и приглашай их в канал:\n" +
		inviteLink.URL

	if err = h.tgUtilsService.SendMessage(message.Chat.ID, text); err != nil {
		log.Error("Failed to SendMessage Invite link [%s] to User [%s, %d]: %v", inviteLink.URL, tgUser.FirstName, tgUser.ID, err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	log.Info("Referral link sent to User [%s, %d] for Channel [%d], InviteLink ID: %d",
		tgUser.FirstName,
		tgUser.ID,
		activeChannel.TelegramID,
		inviteLink.ID,
	)
}

func (h *messageHandler) handleCountCommand(message *tgbotapi.Message) {
	err, _, log, _ := h.core.GetAll()

	// Getting and verifying active channel
	var activeChannel *types.Channel
	if activeChannel, err = h.activeChannelService.Get(); err != nil {
		log.Error("Referral command aborted: Failed to Get Active Channel: %v", err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	} else if activeChannel == nil {
		log.Warn("Referral command aborted: Active Channel is nil.")
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	tgUser := message.From

	// Checking if user is a member
	var userMemberStatus types.MemberStatus
	userMemberStatus, err = h.userService.RequestMemberStatusByTelegramID(tgUser.ID)
	if err != nil {
		log.Error("Referral command aborted: Failed to RequestMemberStatusByTelegramID: %v", err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	if userMemberStatus.NotInChannel() {
		text := "" +
			"Не вижу тебя в подписчиках канала «Медиа Снабжение»!\n\n" +
			"Давай исправим, присоединяйся и к каналу, и к розыгрышу.\n"
		if activeChannel.InviteLink != "" {
			text += activeChannel.InviteLink
		} else {
			log.Warn("No available InviteLink in active channel")
		}

		h.tgUtilsService.SendMessage(message.Chat.ID, text)
		return
	}

	// Adding user to DB, if he's not
	_, err = h.userService.UpdateOrRegister(tgUser)
	if err != nil {
		log.Error("Failed to UpdateOrRegister User [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	// Getting previously created Invite Link
	var inviteLink *types.InviteLink
	inviteLink, err = h.inviteLinkService.GetByRequesterTelegramID(tgUser.ID)
	if err != nil {
		log.Error("Failed to GetByRequester link for user [%s, %d]: %v", tgUser.FirstName, tgUser.ID, err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	// Chec
	if inviteLink == nil {
		text := "" +
			"Упс, у тебя пока нет своей реферальной ссылки!\n\n" +
			"Вводи команду /referal и приглашай коллег в канал. Тогда команда /count покажет, сколько человек пришло по твоей ссылке."
		h.tgUtilsService.SendMessage(message.Chat.ID, text)
		return
	}

	count := inviteLink.UniqueJoins
	switch {
	case count > 0:
		text := "" +
			"По твоей ссылке подписались на канал " + fmt.Sprintf("%d", count) + " человек.\n\n" +
			"Все получается, продолжай делиться своей ссылкой с коллегами!"
		h.tgUtilsService.SendMessage(message.Chat.ID, text)
	case count == 0:
		text := "" +
			"Пока по твоей ссылке никто не подписался на канал.\n\n" +
			"Если у тебя нет ссылки, вводи команду /referral и приглашай коллег в канал. Тогда команда /count покажет, сколько человек пришло по твоей ссылке."
		h.tgUtilsService.SendMessage(message.Chat.ID, text)
	case count < 0:
		log.Error("Less than zero value [%d] on Invite Link. ID: %d, URL: %s", inviteLink.UniqueJoins, inviteLink.ID, inviteLink.URL)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
	}
}

func (h *messageHandler) handleHelpCommand(message *tgbotapi.Message) {
	_, _, log, _ := h.core.GetAll()

	if !h.activeChannelService.IsActive() {
		log.Warn("Help command aborted: No active channel.")
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	text := "" +
		"Доступные команды:\n" +
		"/start - Начать работу\n" +
		"/referral - Получить ссылку-приглашение в канал\n" +
		"/count - Узнать, сколько человек пришло по твоей ссылке\n" +
		"/help - Помощь"

	h.tgUtilsService.SendMessage(message.Chat.ID, text)
}
func (h *messageHandler) handleLogAdmin(message *tgbotapi.Message) {
	log := h.core.GetLogger()

	password := strings.TrimPrefix(message.Text, "/logadmin ")
	if password != "111" {
		h.handleUnknownCommand(message)
		return
	}

	isAdmin, err := h.adminService.IsAdmin(message.From.ID)
	if err != nil {
		log.Error("IsAdmin function for [%s, %d] failed: %v", message.From.FirstName, message.From.ID, err)
		return
	}

	if isAdmin {
		h.tgUtilsService.SendMessage(message.Chat.ID, "Already admin")
		log.Info("User [%s, %d] is already admin", message.From.FirstName, message.From.ID)
		return
	}

	admin, err := h.adminService.Register(message.From)
	if err != nil || admin == nil {
		log.Error("Failed to register user [%s, %d] as admin: %v", message.From.FirstName, message.From.ID, err)
	}

	text := fmt.Sprintf(""+
		"Добавил админа\n"+
		"ID: %d\n"+
		"Telegram ID: %d\n"+
		"First name: %s\n"+
		"Last name: %s\n"+
		"User name: %s\n",
		admin.ID,
		admin.TelegramID,
		admin.FirstName,
		admin.LastName,
		admin.Username,
	)

	h.tgUtilsService.SendMessage(message.Chat.ID, text)
	log.Info(text)
}

func (h *messageHandler) handleRegisterAsActive(message *tgbotapi.Message) {
	log := h.core.GetLogger()

	channelID, err := strconv.Atoi(strings.TrimPrefix(message.Text, "/regact "))
	if err != nil {
		log.Error("Failed to parse Register Active command: %v", err)
		return
	}

	activeChannel, err := h.activeChannelService.Register(int64(channelID))
	if err != nil {
		log.Error("Failed to Register new Active channel [%d]: %v", int64(channelID), err)
		return
	}
	if activeChannel == nil {
		log.Error("Active channel [%d] is nil for some reason")
		return
	}

	log.Info("Active channel registered: ID=%d, Title='%s'",
		activeChannel.ID,
		activeChannel.Title,
	)
	h.tgUtilsService.SendMessage(message.Chat.ID, "Active channel registered")
}

func (h *messageHandler) handleUnknownCommand(message *tgbotapi.Message) {
	_, _, log, _ := h.core.GetAll()

	if !h.activeChannelService.IsActive() {
		log.Warn("Unknown command aborted: No active channel.")
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	text := "" +
		"Неизвестная команда.\n" +
		"Используй /help для списка команд."

	h.tgUtilsService.SendMessage(message.Chat.ID, text)
}

func (h *messageHandler) handleTextMessage(message *tgbotapi.Message) {
	_, _, log, _ := h.core.GetAll()

	if !h.activeChannelService.IsActive() {
		log.Warn("Text message handling aborted: No active channel.")
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	h.handleUnknownMessage(message)
}

func (h *messageHandler) handleUnknownMessage(message *tgbotapi.Message) {
	text := "" +
		"Я тебя не понимаю!\n" +
		"Воспользуйся командой /help, если запутался."
	h.tgUtilsService.SendMessage(message.Chat.ID, text)
}
