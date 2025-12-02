package receivers

import (
	"fmt"
	"referral-bot/internal/config"
	"referral-bot/internal/core"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Poller struct {
	*receiverBase
	telegramUpdatesChannel tgbotapi.UpdatesChannel
}

func NewPoller(core *core.Core) (*Poller, error) {
	base, err := newReceiverBase(core)
	if err != nil {
		return nil, fmt.Errorf("failed to create baseReceiver: %w", err)
	}

	config := &core.GetConfig().Receiver
	if err := verifyPollerConfig(config); err != nil {
		return nil, fmt.Errorf("wrong poller config: %w", err)
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

	if p.updateHandler == nil {
		return fmt.Errorf("update handler cannot be nil")
	}

	if err := p.setupReceiverBase(); err != nil {
		return fmt.Errorf("failed to setup receiver base: %w", err)
	}

	if err := p.setupTelegramUpdatesChannel(); err != nil {
		return fmt.Errorf("failed to setup telegram poller: %w", err)
	}

	go p.runPoller()

	go p.processUpdatesFromBuffer()

	p.core.GetLogger().Info("Poller started")
	return nil
}

func (p *Poller) setupTelegramUpdatesChannel() error {
	config := p.core.GetConfig().Receiver
	updateConfig := tgbotapi.NewUpdate(config.Poller.Offset)
	updateConfig.Timeout = config.Poller.Timeout
	updateConfig.AllowedUpdates = config.AllowedUpdates

	p.telegramUpdatesChannel = p.core.GetBotAPI().GetUpdatesChan(updateConfig)

	p.core.GetLogger().Info("Telegram poller configured")

	return nil
}

func (p *Poller) runPoller() {
	log := p.core.GetLogger()
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
		return fmt.Errorf("failed to stop receiver base: %w", err)
	}

	p.core.GetLogger().Info("Poller base stopped gracefully")
	return nil
}

func (p *Poller) GetType() string {
	return "poller"
}
