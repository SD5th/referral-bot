package interfaces

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type UpdateHandler interface {
	HandleUpdate(tgbotapi.Update) error
}
