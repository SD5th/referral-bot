package handlers

import (
	"log"
	"referral-bot/internal/core"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (u *UpdateHandler) handleMessage(message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	if message.IsCommand() {
		handleCommand(u.core, message)
		return
	}
	handleTextMessage(u.core, message)
}

func handleCommand(core *core.Core, message *tgbotapi.Message) {
	command := message.Command()

	switch command {
	case "start":
		handleStartCommand(core, message)
	/*
		case "referral":
			handleReferralCommand(core, message)
		case "db":
			handleDatabaseCommand(core, message)
	*/
	case "help":
		handleHelpCommand(core, message)
	default:
		handleUnknownCommand(core, message)
	}
}

func handleStartCommand(core *core.Core, message *tgbotapi.Message) {
	text := `Привет! Я бот для розыгрыша.`

	SendMessage(core.GetBotAPI(), message.Chat.ID, text)
}

/*
func handleReferralCommand(core *core.Core, message *tgbotapi.Message) {

	referral, _ := tgutils.GenerateInviteLink(bot, -1003485818455, strconv.FormatInt(message.From.ID, 10))
	text := `Привет! Я бот для розыгрыша.
Вот твоя ссылка:` + referral

	tgutils.SendMessage(bot, message.Chat.ID, text)
}

func handleDatabaseCommand(core *core.Core, message *tgbotapi.Message) {
	storage.NewDatabase()
}
*/

func handleHelpCommand(core *core.Core, message *tgbotapi.Message) {
	text := `Доступные команды:
/start - Начать работу
/referral - Сгенерировать реферальную ссылку
/db - database
/help - Помощь`

	SendMessage(core.GetBotAPI(), message.Chat.ID, text)
}

func handleUnknownCommand(core *core.Core, message *tgbotapi.Message) {
	text := `Неизвестная команда.
Используй /help для списка команд.`

	SendMessage(core.GetBotAPI(), message.Chat.ID, text)
}

func handleTextMessage(core *core.Core, message *tgbotapi.Message) {
	handleUnknownMessage(core, message)
}

func handleUnknownMessage(core *core.Core, message *tgbotapi.Message) {
	text := "" +
		"Ты мне сказал:\n" +
		message.Text + "\n\n" +
		"Я тебя не понимаю!\n" +
		"Воспользуйся командой /help, если запутался."
	SendMessage(core.GetBotAPI(), message.Chat.ID, text)
}

// /////////
func SendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки: %v", err)
	}
}
