package handlers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Логируем полученное сообщение
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	// Создаем ответное сообщение (эхо)
	msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)

	// Отправляем сообщение
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
