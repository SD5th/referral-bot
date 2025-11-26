package core

import (
	"referral-bot/internal/config"
	"referral-bot/internal/interfaces"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Core struct {
	botAPI *tgbotapi.BotAPI
	config *config.Config
	logger interfaces.LoggerInterface
}

func NewCore() *Core {
	return &Core{
		botAPI: nil,
		config: nil,
		logger: nil,
	}
}

func (c *Core) GetBotAPI() *tgbotapi.BotAPI {
	return c.botAPI
}

func (c *Core) GetConfig() *config.Config {
	return c.config
}

func (c *Core) GetLogger() interfaces.LoggerInterface {
	return c.logger
}

func (c *Core) SetBotAPI(botAPI *tgbotapi.BotAPI) {
	c.botAPI = botAPI
}

func (c *Core) SetConfig(config *config.Config) {
	c.config = config
}

func (c *Core) SetLogger(logger interfaces.LoggerInterface) {
	c.logger = logger
}
