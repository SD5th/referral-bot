package bot

import (
	"fmt"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API            *tgbotapi.BotAPI
	Config         *types.BotConfig
	UpdateReceiver *UpdateReceiver
}

func NewBot() (*Bot, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}
	pubC := config.PublicConfig
	privC := config.PrivateConfig

	api, err := tgbotapi.NewBotAPI(privC.Token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		API:            api,
		Config:         config,
		UpdateReceiver: nil,
	}

	bot.API.Debug = pubC.Debug

	var updateReceiver UpdateReceiver
	switch pubC.UpdateReceiverType {
	case "poller":
		updateReceiver, err = NewPoller(bot)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Неизвестный UpdateReceiverType")
	}
	bot.UpdateReceiver = &updateReceiver

	return bot, nil
}

func (bot *Bot) GetPublicConfig() *types.BotPublicConfig {
	return bot.Config.PublicConfig
}

func (bot *Bot) GetAPI() *tgbotapi.BotAPI {
	return bot.API
}

func (bot *Bot) StartReceiver() error {
	return (*bot.UpdateReceiver).start()
}

func (bot *Bot) StopReceiver() error {
	return (*bot.UpdateReceiver).stop()
}

func (bot *Bot) IsReceiving() bool {
	return (*bot.UpdateReceiver).isRunning()
}
