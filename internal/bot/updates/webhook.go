package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"referral-bot/internal/config"
	"referral-bot/internal/types"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Webhook struct {
	*receiverBase
	server *http.Server
}

func NewWebhook(bot types.BotContext, config *config.ReceiverConfig) (*Webhook, error) {
	base, err := newReceiverBase(bot, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create baseReceiver: %v", err)
	}

	if err := verifyWebhookConfig(config); err != nil {
		return nil, fmt.Errorf("wrong webhook config: %v", err)
	}

	return &Webhook{
		receiverBase: base,
		server:       nil,
	}, nil
}

func verifyWebhookConfig(config *config.ReceiverConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Webhook.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	// TODO: ДОБАВИТЬ ПРОВЕРКУ IP

	// TODO: ДОБАВИТЬ ПРОВЕРКУ Port

	// TODO: ДОБАВИТЬ ПРОВЕРКУ CertFile

	// TODO: ДОБАВИТЬ ПРОВЕРКУ KeyFile

	// TODO: ДОБАВИТЬ ПРОВЕРКУ ReadTimeout

	return nil
}

func (w *Webhook) Start() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if err := w.setupReceiverBase(); err != nil {
		return fmt.Errorf("failed to setup receiver base: %v", err)
	}

	if err := w.setupTelegramWebhook(); err != nil {
		return fmt.Errorf("failed to setup telegram webhook: %v", err)
	}

	if err := w.setupServer(); err != nil {
		return fmt.Errorf("failed to setup webhook server: %v", err)
	}

	go w.runServer()

	go w.processUpdatesFromBuffer()

	w.bot.GetLogger().Info("Webhook server started on %s:%s", w.config.Webhook.IP, w.config.Webhook.Port)

	return nil
}

func (w *Webhook) setupTelegramWebhook() error {
	webhookConfig, err := tgbotapi.NewWebhook(w.config.Webhook.URL)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %v", err)
	}

	webhookConfig.AllowedUpdates = w.config.AllowedUpdates
	webhookConfig.Certificate = tgbotapi.FilePath(w.config.Webhook.CertFile)

	_, err = w.bot.GetAPI().Request(webhookConfig)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %v", err)
	}

	w.bot.GetLogger().Info("Telegram webhook configured to: %s", w.config.Webhook.URL)
	return nil
}

func (w *Webhook) setupServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", w.webhookHandler)
	mux.HandleFunc("/health", w.healthHandler)

	w.server = &http.Server{
		Addr:         w.config.Webhook.IP + ":" + w.config.Webhook.Port,
		Handler:      mux,
		ReadTimeout:  time.Duration(w.config.Webhook.ReadTimeout) * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return nil
}

func (w *Webhook) runServer() {
	serverErr := make(chan error, 1)

	go func() {
		err := w.server.ListenAndServeTLS(w.config.Webhook.CertFile, w.config.Webhook.KeyFile)
		serverErr <- err
	}()

	select {
	case <-w.ctx.Done():
		w.shutdownServer()
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			w.bot.GetLogger().Warn("Webhook server error: %v", err)
		}
	}
}

func (w *Webhook) webhookHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var update tgbotapi.Update
	if err := json.NewDecoder(request.Body).Decode(&update); err != nil {
		w.bot.GetLogger().Warn("Error decoding update: %v", err)
		http.Error(writer, "Bad request", http.StatusBadRequest)
		return
	}

	if err := w.sendUpdateToBuffer(update); err != nil {
		w.bot.GetLogger().Warn("Send update failed")
	}

	writer.WriteHeader(http.StatusOK)
}

func (w *Webhook) healthHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"receiver":  "webhook",
	})
}

func (w *Webhook) shutdownServer() {
	if w.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := w.server.Shutdown(ctx); err != nil {
			w.bot.GetLogger().Warn("Webhook server shutdown error: %v", err)
		} else {
			w.bot.GetLogger().Info("Webhook server stopped gracefully")
		}
	}
}

func (w *Webhook) Stop() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.running {
		return nil
	}

	w.shutdownServer()

	if err := w.removeTelegramWebhook(); err != nil {
		w.bot.GetLogger().Warn("Warning: failed to remove telegram webhook: %v", err)
	}

	if err := w.stopReceiverBase(); err != nil {
		return fmt.Errorf("failed to stop receiver base: %v", err)
	}

	w.bot.GetLogger().Info("Webhook stopped gracefully")

	return nil
}

func (w *Webhook) removeTelegramWebhook() error {
	_, err := w.bot.GetAPI().Request(tgbotapi.DeleteWebhookConfig{
		DropPendingUpdates: false,
	})
	return err
}

func (w *Webhook) GetType() string {
	return "webhook"
}
