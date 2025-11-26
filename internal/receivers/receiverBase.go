package receivers

import (
	"context"
	"fmt"
	"referral-bot/internal/config"
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type receiverBase struct {
	core          *core.Core
	updateHandler interfaces.UpdateHandlerInterface

	running       bool
	ctx           context.Context
	cancel        context.CancelFunc
	mutex         sync.RWMutex
	updatesBuffer chan tgbotapi.Update
}

func newReceiverBase(core *core.Core) (*receiverBase, error) {
	if core == nil {
		return nil, fmt.Errorf("core cannot be nil")
	}

	config := &core.GetConfig().Receiver
	if err := verifyReceiverBaseConfig(config); err != nil {
		return nil, fmt.Errorf("wrong receiver config: %v", err)
	}

	return &receiverBase{
		core:          core,
		updateHandler: nil,

		running:       false,
		ctx:           nil,
		cancel:        nil,
		mutex:         sync.RWMutex{},
		updatesBuffer: nil,
	}, nil
}

func (b *receiverBase) SetUpdateHandler(updateHandler interfaces.UpdateHandlerInterface) error {
	if updateHandler == nil {
		return fmt.Errorf("update handler cannot be nil")
	}

	b.updateHandler = updateHandler

	return nil
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

	bufferSize := b.core.GetConfig().Receiver.BufferSize
	b.updatesBuffer = make(chan tgbotapi.Update, bufferSize)

	return nil
}

func (b *receiverBase) sendUpdateToBuffer(update tgbotapi.Update) error {
	if b.updatesBuffer == nil {
		return fmt.Errorf("update buffer channel is not initialized")
	}

	log := b.core.GetLogger()

	defer func() {
		if r := recover(); r != nil {
			log.Warn("Recovery from sendUpdate panic: %v", r)
		}
	}()

	select {
	case b.updatesBuffer <- update:
		return nil
	default:
		log.Warn("Update buffer full, waiting with timeout...")
	}

	select {
	case b.updatesBuffer <- update:
		log.Info("Update %d sent to buffer after waiting", update.UpdateID)
		return nil
	case <-time.After(5 * time.Second):
		log.Warn("Timeout sending update - buffer blocked for 5 seconds")
		return fmt.Errorf("timeout sending update")
	}
}

func (b *receiverBase) processUpdatesFromBuffer() {
	log := b.core.GetLogger()
	if b.updateHandler == nil {
		log.Fatal("Update handler cannot be nil")
	}

	defer func() {
		if r := recover(); r != nil {
			log.Warn("Recovered from panic in processUpdatesFromBuffer: %v", r)
		}
	}()

	for {
		select {
		case update, ok := <-b.updatesBuffer:
			if !ok {
				log.Info("update buffer closed")
				return
			}

			b.updateHandler.HandleUpdate(update)

		case <-b.ctx.Done():
			log.Info("update processing stopped by context")
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

	b.core.GetLogger().Info("Receiver base stopped gracefully")
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
