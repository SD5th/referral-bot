package services

import (
	"fmt"
	"referral-bot/internal/core"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TGUtilsService struct {
	core *core.Core
}

func NewTGUtilsService(core *core.Core) *TGUtilsService {
	return &TGUtilsService{core: core}
}

func (s *TGUtilsService) SendMessage(chatID int64, text string) error {
	err, botAPI, log, _ := s.core.GetAll()
	if err != nil {
		log.Error("Failed to get bot API: %v", err)
		return nil
	}

	if text == "" {
		log.Warn("Attempted to send empty message to chat %d", chatID)
		return nil
	}

	// Создаем сообщение
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	// Отправляем сообщение
	_, err = botAPI.Send(msg)
	if err != nil {
		log.Error("Failed to send message to chat %d: %v", chatID, err)
		return fmt.Errorf("Failed to send message to chat %d: %v", chatID, err)
	}

	log.Info("Message sent to chat %d: %v", chatID, text)
	return nil
}

func (s *TGUtilsService) SendTryAgainLater(chatID int64) error {
	return s.SendMessage(chatID, "Сервис временно недоступен. Попробуйте позже.")
}
