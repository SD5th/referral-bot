package updates

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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
	webhookToken string
	server       *http.Server
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
		webhookToken: "",
	}, nil
}

func verifyWebhookConfig(config *config.ReceiverConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Webhook.ServerIP == "" {
		return fmt.Errorf("server IP is required")
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

	if err := w.generateWebhookToken(); err != nil {
		return fmt.Errorf("failed to generate webhook token: %v", err)
	}

	if err := w.setupServer(); err != nil {
		return fmt.Errorf("failed to setup webhook server: %v", err)
	}

	if err := w.setupTelegramWebhook(); err != nil {
		return fmt.Errorf("failed to setup telegram webhook: %v", err)
	}

	go w.runServer()

	go w.processUpdatesFromBuffer()

	w.bot.GetLogger().Info("Webhook server started on %s:%s", w.config.Webhook.ServerIP, w.config.Webhook.Port)

	return nil
}

func (w *Webhook) generateWebhookToken() error {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Errorf("failed to generate random bytes: %v", err)
	}

	w.webhookToken = base64.URLEncoding.EncodeToString(bytes)

	return nil
}

func (w *Webhook) setupServer() error {
	mux := http.NewServeMux()
	webhookPath := "/webhook/" + w.webhookToken
	mux.HandleFunc(webhookPath, w.webhookHandler)

	w.server = &http.Server{
		Addr:         w.config.Webhook.ServerIP + ":" + w.config.Webhook.Port,
		Handler:      mux,
		ReadTimeout:  time.Duration(w.config.Webhook.ReadTimeout) * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return nil
}

func (w *Webhook) setupTelegramWebhook() error {
	webhookURL := "https://" + w.config.Webhook.ServerIP + "/webhook/" + w.webhookToken
	webhookConfig, err := tgbotapi.NewWebhookWithCert(webhookURL, tgbotapi.FilePath(w.config.Webhook.CertFile))
	if err != nil {
		return fmt.Errorf("failed to create webhook: %v", err)
	}

	webhookConfig.AllowedUpdates = w.config.AllowedUpdates

	_, err = w.bot.GetAPI().Request(webhookConfig)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %v", err)
	}

	w.bot.GetLogger().Info("Telegram webhook configured to: %s", webhookURL)
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
