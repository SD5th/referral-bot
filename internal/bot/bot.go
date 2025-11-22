package bot

import (
	"fmt"
	"referral-bot/internal/bot/updates"
	"referral-bot/internal/config"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API            *tgbotapi.BotAPI
	Config         *config.BotConfig
	Logger         types.LoggerContext
	updateReceiver *updates.UpdateReceiver
}

func NewBot(config *config.BotConfig, logger types.LoggerContext) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		API:            api,
		Config:         config,
		Logger:         logger,
		updateReceiver: nil,
	}

	bot.API.Debug = config.Debug

	var updateReceiver updates.UpdateReceiver
	switch config.Receiver.Type {
	case "poller":
		updateReceiver, err = updates.NewPoller(bot, &config.Receiver)
		if err != nil {
			return nil, err
		}
	case "webhook":
		updateReceiver, err = updates.NewWebhook(bot, &config.Receiver)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown UpdateReceiverType")
	}
	bot.updateReceiver = &updateReceiver

	return bot, nil
}

func (bot *Bot) GetConfig() *config.BotConfig {
	return bot.Config
}

func (bot *Bot) GetLogger() types.LoggerContext {
	return bot.Logger
}

func (bot *Bot) GetAPI() *tgbotapi.BotAPI {
	return bot.API
}

func (bot *Bot) StartReceiver() error {
	return (*bot.updateReceiver).Start()
}

func (bot *Bot) StopReceiver() error {
	return (*bot.updateReceiver).Stop()
}

func (bot *Bot) IsReceiving() bool {
	return (*bot.updateReceiver).IsRunning()
}
