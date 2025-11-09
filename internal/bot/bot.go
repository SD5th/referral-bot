package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API            *tgbotapi.BotAPI
	config         *Config
	UpdateReceiver *UpdateReceiver
}

func NewBot() (*Bot, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, err
	}

	var updateReceiver UpdateReceiver
	switch config.UpdateReceiverType {
	case "poller":
		updateReceiver, err = NewPoller(api)
		if err != nil {
			return nil, err
		}
	default:
		log.Fatalf("Неизвестный UpdateReceiverType")
	}

	return &Bot{
		API:            api,
		config:         config,
		UpdateReceiver: &updateReceiver,
	}, nil
}
