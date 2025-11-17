package updates

import (
	"context"
	"fmt"
	"log"
	"referral-bot/internal/config"
	"referral-bot/internal/types"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Poller struct {
	bot     types.BotContext
	running *bool
	cancel  context.CancelFunc
	mutex   *sync.RWMutex
	config  *config.ReceiverConfig
}

func NewPoller(bot types.BotContext, config *config.ReceiverConfig) (*Poller, error) {
	if bot == nil {
		return nil, fmt.Errorf("api cannot be nil")
	}

	running := false
	return &Poller{
		bot:     bot,
		running: &running,
		cancel:  nil,
		mutex:   &sync.RWMutex{},
		config:  config,
	}, nil
}

func (p *Poller) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if *p.running {
		return fmt.Errorf("poller is already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	*p.running = true

	log.Printf("Starting poller")

	go p.run(ctx)

	return nil
}

func (p *Poller) run(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	u.AllowedUpdates = p.config.AllowedUpdates

	updates := p.bot.GetAPI().GetUpdatesChan(u)

	for {
		select {
		case update, ok := <-updates:
			if !ok {
				log.Println("Update channel closed")
				return
			}

			go handleUpdate(p.bot, update)

		case <-ctx.Done():
			log.Println("Poller stopped by context")
			return
		}
	}
}

func (p *Poller) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !(*p.running) {
		return nil
	}

	if p.cancel != nil {
		p.cancel()
	}

	*p.running = false
	log.Println("Poller stopped gracefully")

	return nil
}

func (p *Poller) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return *p.running
}

func (p *Poller) GetType() string {
	return "poller"
}
