package tgutils

import (
	"log"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendMessage(bot types.BotContext, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	_, err := bot.GetAPI().Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки: %v", err)
	}
}
