package types

import (
	"referral-bot/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotContext interface {
	GetConfig() *config.BotConfig
	GetAPI() *tgbotapi.BotAPI
	StartReceiver() error
	StopReceiver() error
	IsReceiving() bool
}
