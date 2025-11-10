package handlers

import (
	"log"
	"referral-bot/internal/bot/tgutils"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessage(bot types.BotContext, message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	if message.IsCommand() {
		handleCommand(bot, message)
		return
	}
	handleTextMessage(bot, message)
}

func handleCommand(bot types.BotContext, message *tgbotapi.Message) {
	command := message.Command()

	switch command {
	case "start":
		handleStartCommand(bot, message)
	case "help":
		handleHelpCommand(bot, message)
	default:
		handleUnknownCommand(bot, message)
	}
}

func handleStartCommand(bot types.BotContext, message *tgbotapi.Message) {
	text := `Привет! Я бот для розыгрыша.`

	tgutils.SendMessage(bot, message.Chat.ID, text)
}

func handleHelpCommand(bot types.BotContext, message *tgbotapi.Message) {
	text := `Доступные команды:
/start - Начать работу
/help - Помощь`

	tgutils.SendMessage(bot, message.Chat.ID, text)
}

func handleUnknownCommand(bot types.BotContext, message *tgbotapi.Message) {
	text := `Неизвестная команда.
Используй /help для списка команд.`

	tgutils.SendMessage(bot, message.Chat.ID, text)
}

func handleTextMessage(bot types.BotContext, message *tgbotapi.Message) {
	handleUnknownMessage(bot, message)
}

// handleUnknownMessage обрабатывает непонятные сообщения
func handleUnknownMessage(bot types.BotContext, message *tgbotapi.Message) {
	text := `Ты мне сказал:
		` + message.Text + `
		Я тебя не понимаю!`

	tgutils.SendMessage(bot, message.Chat.ID, text)
}
