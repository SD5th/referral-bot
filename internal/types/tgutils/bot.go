package types

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type BotConfig struct {
	PublicConfig  *BotPublicConfig
	PrivateConfig *BotPrivateConfig
}

type BotPublicConfig struct {
	Debug              bool   `json:"debug"`
	UpdateReceiverType string `json:"updateReceiverType"`
}

type BotPrivateConfig struct {
	Token string `json:"token"`
}

type BotContext interface {
	GetPublicConfig() *BotPublicConfig
	GetAPI() *tgbotapi.BotAPI
	StartReceiver() error
	StopReceiver() error
	IsReceiving() bool
}
