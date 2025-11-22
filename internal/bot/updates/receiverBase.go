package updates

import (
	"context"
	"fmt"
	"referral-bot/internal/config"
	"referral-bot/internal/types"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type receiverBase struct {
	bot           types.BotContext
	running       bool
	ctx           context.Context
	cancel        context.CancelFunc
	mutex         sync.RWMutex
	updatesBuffer chan tgbotapi.Update
	config        *config.ReceiverConfig
}

func newReceiverBase(bot types.BotContext, config *config.ReceiverConfig) (*receiverBase, error) {
	if bot == nil {
		return nil, fmt.Errorf("api cannot be nil")
	}
	if err := verifyReceiverBaseConfig(config); err != nil {
		return nil, fmt.Errorf("wrong receiver config: %v", err)
	}

	return &receiverBase{
		bot:           bot,
		running:       false,
		mutex:         sync.RWMutex{},
		updatesBuffer: nil,
		config:        config,
	}, nil
}

func verifyReceiverBaseConfig(config *config.ReceiverConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.BufferSize <= 0 {
		return fmt.Errorf("buffer size must be positive")
	}

	// TODO: ДОБАВИТЬ ПРОВЕРКУ ALLOWED UPDATES

	return nil
}

func (b *receiverBase) IsRunning() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.running
}

// MUST be called under mutex lock from Start() method.
// NOT safe for concurrent use without external synchronization.
func (b *receiverBase) setupReceiverBase() error {
	if b.running {
		return fmt.Errorf("receiver is already running")
	}

	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}

	if err := b.openUpdatesBuffer(); err != nil {
		return fmt.Errorf("failed to open updates buffer: %v", err)
	}
	b.running = true

	b.ctx, b.cancel = context.WithCancel(context.Background())
	return nil
}

func (b *receiverBase) openUpdatesBuffer() error {
	if b.updatesBuffer != nil {
		return fmt.Errorf("update buffer is already opened")
	}
	b.updatesBuffer = make(chan tgbotapi.Update, b.config.BufferSize)

	return nil
}

func (b *receiverBase) sendUpdateToBuffer(update tgbotapi.Update) error {
	if b.updatesBuffer == nil {
		return fmt.Errorf("update channel is not initialized")
	}

	defer func() {
		if r := recover(); r != nil {
			b.bot.GetLogger().Warn("Recovery from sendUpdate panic: %v", r)
		}
	}()

	select {
	case b.updatesBuffer <- update:
		return nil
	default:
		b.bot.GetLogger().Warn("Update buffer full, waiting with timeout...")
	}

	select {
	case b.updatesBuffer <- update:
		b.bot.GetLogger().Info("Update %d sent to buffer after waiting", update.UpdateID)
		return nil
	case <-time.After(5 * time.Second):
		b.bot.GetLogger().Warn("Timeout sending update - buffer blocked for 5 seconds")
		return fmt.Errorf("timeout sending update")
	}
}

func (b *receiverBase) processUpdatesFromBuffer() {
	defer func() {
		if r := recover(); r != nil {
			b.bot.GetLogger().Warn("Recovered from panic in processUpdatesFromBuffer: %v", r)
		}
	}()

	for {
		select {
		case update, ok := <-b.updatesBuffer:
			if !ok {
				b.bot.GetLogger().Info("update buffer closed")
				return
			}

			handleUpdate(b.bot, update)

		case <-b.ctx.Done():
			b.bot.GetLogger().Info("update processing stopped by context")
			return
		}
	}
}

// MUST be called under mutex lock from Stop() method.
// NOT safe for concurrent use without external synchronization.
func (b *receiverBase) stopReceiverBase() error {
	if b.cancel == nil {
		return fmt.Errorf("cancel function is nil")
	}

	if err := b.closeUpdatesBuffer(); err != nil {
		return fmt.Errorf("failed to close updates buffer: %v", err)
	}

	b.cancel()
	b.cancel = nil

	b.running = false

	b.bot.GetLogger().Info("Receiver base stopped gracefully")
	return nil
}

func (b *receiverBase) closeUpdatesBuffer() error {
	if b.updatesBuffer == nil {
		return nil
	}

	close(b.updatesBuffer)
	b.updatesBuffer = nil

	return nil
}
