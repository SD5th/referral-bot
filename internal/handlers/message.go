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

	if !h.activeChannelService.IsActive() {
		log.Warn("Referral command aborted: No active channel.")
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	// Getting active channel
	var activeChannel *types.Channel
	if activeChannel, err = h.activeChannelService.Get(); err != nil {
		log.Error("Failed to Get Active Channel: %v", err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	} else if activeChannel == nil {
		log.Warn("Active Channel is nil.")
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	// Getting user
	user, err := h.userService.GetOrUpdateFromMessage(message)
	if err != nil {
		log.Error("Failed to GetOrUpdateFromMessage User [%s, %d]: %v", message.From.FirstName, message.From.ID, err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}
	if user == nil {
		user, err = h.userService.RegisterInDB(message.From)
		if err != nil {
			log.Error("Failed to RegisterInDB User [%s, %d]: %v", message.From.FirstName, message.From.ID, err)
			h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
			return
		}
	}

	// Checking if user is a member
	if !user.IsMember() {
		var text string
		if activeChannel.InviteLink != "" {
			text = "" +
				"Не вижу тебя в подписчиках канала «Медиа Снабжение»!\n\n" +
				"Давай исправим, присоединяйся и к каналу, и к розыгрышу:\n" +
				activeChannel.InviteLink
		} else {
			log.Warn("No available InviteLink in active channel")
			text = "" +
				"Не вижу тебя в подписчиках канала «Медиа Снабжение»!\n\n" +
				"Давай исправим, присоединяйся и к каналу, и к розыгрышу."
		}

		h.tgUtilsService.SendMessage(message.Chat.ID, text)
		return
	}

	// Getting previously created Invite Link
	inviteLink, err := h.inviteLinkService.GetByRequester(user)
	if err != nil {
		log.Error("Failed to GetByRequester User [%s, %d]: %v", message.From.FirstName, message.From.ID, err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	// Creating new Invite Link
	if inviteLink == nil {
		inviteLink, err = h.inviteLinkService.CreateForRequester(user)
		if err != nil {
			log.Error("Failed to CreateInviteLink for User [%s, %d]: %v", user.FirstName, user.TelegramID, err)
			h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
			return
		}
	}

	// Sending Link
	text := "" +
		"Пересылай эту ссылку коллегам и приглашай их в канал:\n" +
		inviteLink.URL

	if err = h.tgUtilsService.SendMessage(message.Chat.ID, text); err != nil {
		log.Error("Failed to Send Invite link [%s] to User [%s, %d]: %v", inviteLink.URL, user.FirstName, user.TelegramID, err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	log.Info("Referral link sent to User [%s, %d] for Channel [%d], InviteLink ID: %d",
		user.FirstName,
		user.TelegramID,
		activeChannel.TelegramID,
		inviteLink.ID,
	)
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

func (h *messageHandler) handleCountCommand(message *tgbotapi.Message) {
	err, _, log, _ := h.core.GetAll()

	if !h.activeChannelService.IsActive() {
		log.Warn("Count command aborted: No active channel.")
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	// Getting active channel
	var activeChannel *types.Channel
	if activeChannel, err = h.activeChannelService.Get(); err != nil {
		log.Error("Failed to Get Active Channel: %v", err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	} else if activeChannel == nil {
		log.Warn("Active Channel is nil.")
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	// Getting user
	user, err := h.userService.GetOrUpdateFromMessage(message)
	if err != nil {
		log.Error("Failed to GetOrUpdateFromMessage User [%s, %d]: %v", message.From.FirstName, message.From.ID, err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}
	if user == nil {
		user, err = h.userService.RegisterInDB(message.From)
		if err != nil {
			log.Error("Failed to RegisterInDB User [%s, %d]: %v", message.From.FirstName, message.From.ID, err)
			h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
			return
		}
	}

	// Checking if user is a member
	if !user.IsMember() {
		var text string
		if activeChannel.InviteLink != "" {
			text = "" +
				"Не вижу тебя в подписчиках канала «Медиа Снабжение»!\n\n" +
				"Давай исправим, присоединяйся и к каналу, и к розыгрышу:\n" +
				activeChannel.InviteLink
		} else {
			log.Warn("No available InviteLink in active channel")
			text = "" +
				"Не вижу тебя в подписчиках канала «Медиа Снабжение»!\n\n" +
				"Давай исправим, присоединяйся и к каналу, и к розыгрышу."
		}

		h.tgUtilsService.SendMessage(message.Chat.ID, text)
		return
	}

	// Getting previously created Invite Link
	inviteLink, err := h.inviteLinkService.GetByRequester(user)
	if err != nil {
		log.Error("Failed to GetByRequester User [%s, %d]: %v", message.From.FirstName, message.From.ID, err)
		h.tgUtilsService.SendTryAgainLater(message.Chat.ID)
		return
	}

	if inviteLink == nil {
		text := "" +
			"Упс, у тебя пока нет своей реферальной ссылки!!\n\n" +
			"Вводи команду /referal и приглашай коллег в канал. Тогда команда /count покажет, сколько человек пришло по твоей ссылке."
		h.tgUtilsService.SendMessage(message.Chat.ID, text)
		return
	}

	if inviteLink.UniqueJoins <= 0 {
		text := "" +
			"Пока по твоей ссылке никто не подписался на канал.\n\n" +
			"Если у тебя нет ссылки, вводи команду /referral и приглашай коллег в канал. Тогда команда /count покажет, сколько человек пришло по твоей ссылке."
		h.tgUtilsService.SendMessage(message.Chat.ID, text)
		return
	}

	inviteLinkCount := fmt.Sprintf("%d", inviteLink.UniqueJoins)
	text := "" +
		"По твоей ссылке подписались на канал " + inviteLinkCount + " человек.\n\n" +
		"Все получается, продолжай делиться своей ссылкой с коллегами!"
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

/*
	func (h *messageHandler) handleLeaveCommand(message *tgbotapi.Message) {
		if !h.activeChannelAvailable() {
			h.messageService.SendMessage(message.Chat.ID, "Активный канал не настроен. Обратитесь к администратору.")
			return
		}
		text := "" +
			"triggered leave command"

		// написать leave команду
		h.messageService.SendMessage(message.Chat.ID, text)
	}
*/

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
