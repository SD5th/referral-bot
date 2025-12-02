package bot

import (
	"fmt"
	"referral-bot/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SetupBotAPI(config *config.BotAPIConfig) (*tgbotapi.BotAPI, error) {
	api, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("creating Bot API error: %w", err)
	}

	api.Debug = config.Debug

	return api, nil
}
