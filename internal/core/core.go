package core

import (
	"fmt"
	"referral-bot/internal/config"
	"referral-bot/internal/interfaces"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Core struct {
	botAPI *tgbotapi.BotAPI
	logger interfaces.Logger
	config *config.Config
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

func (c *Core) GetLogger() interfaces.Logger {
	return c.logger
}

func (c *Core) GetConfig() *config.Config {
	return c.config
}

func (c *Core) GetAll() (error, *tgbotapi.BotAPI, interfaces.Logger, *config.Config) {
	if c.botAPI == nil {
		return fmt.Errorf("core: botAPI cannot be nil"), nil, nil, nil
	}
	if c.logger == nil {
		return fmt.Errorf("core: logger cannot be nil"), nil, nil, nil
	}
	if c.config == nil {
		return fmt.Errorf("core: config cannot be nil"), nil, nil, nil
	}

	return nil, c.botAPI, c.logger, c.config
}

func (c *Core) SetBotAPI(botAPI *tgbotapi.BotAPI) {
	c.botAPI = botAPI
}

func (c *Core) SetConfig(config *config.Config) {
	c.config = config
}

func (c *Core) SetLogger(logger interfaces.Logger) {
	c.logger = logger
}
