package updates

import (
	"fmt"
	"referral-bot/internal/config"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Poller struct {
	*receiverBase
	telegramUpdatesChannel tgbotapi.UpdatesChannel
}

func NewPoller(bot types.BotContext, config *config.ReceiverConfig) (*Poller, error) {
	base, err := newReceiverBase(bot, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create baseReceiver: %v", err)
	}

	if err := verifyPollerConfig(config); err != nil {
		return nil, fmt.Errorf("wrong poller config: %v", err)
	}

	return &Poller{
		receiverBase: base,
	}, nil
}

func verifyPollerConfig(config *config.ReceiverConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// TODO: ДОБАВИТЬ ПРОВЕРКУ OFFSET

	// TODO: ДОБАВИТЬ ПРОВЕРКУ TIMEOUT

	return nil
}

func (p *Poller) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.setupReceiverBase(); err != nil {
		return fmt.Errorf("failed to setup receiver base: %v", err)
	}

	if err := p.setupTelegramUpdatesChannel(); err != nil {
		return fmt.Errorf("failed to setup telegram poller: %v", err)
	}

	go p.runPoller()

	go p.processUpdatesFromBuffer()

	p.bot.GetLogger().Info("Poller started")
	return nil
}

func (p *Poller) setupTelegramUpdatesChannel() error {
	updateConfig := tgbotapi.NewUpdate(p.config.Poller.Offset)
	updateConfig.Timeout = p.config.Poller.Timeout
	updateConfig.AllowedUpdates = p.config.AllowedUpdates

	p.telegramUpdatesChannel = p.bot.GetAPI().GetUpdatesChan(updateConfig)

	p.bot.GetLogger().Info("Telegram poller configured")

	return nil
}

func (p *Poller) runPoller() {
	log := p.bot.GetLogger()
	log.Info("Starting poller update forwarding...")
	for {
		select {
		case update, ok := <-p.telegramUpdatesChannel:
			if !ok {
				log.Info("Official poller channel closed")
				return
			}

			if err := p.sendUpdateToBuffer(update); err != nil {
				log.Warn("Send update failed")
			}

		case <-p.ctx.Done():
			log.Info("Poller forwarding stopped by context")
			return
		}
	}
}

func (p *Poller) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.running {
		return nil
	}

	if err := p.stopReceiverBase(); err != nil {
		return fmt.Errorf("failed to stop receiver base: %v", err)
	}

	p.bot.GetLogger().Info("Poller base stopped gracefully")
	return nil
}

func (p *Poller) GetType() string {
	return "poller"
}
