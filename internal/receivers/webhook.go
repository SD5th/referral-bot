package receivers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"referral-bot/internal/config"
	"referral-bot/internal/core"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Webhook struct {
	*receiverBase
	server       *http.Server
	webhookToken string
}

func NewWebhook(core *core.Core) (*Webhook, error) {
	base, err := newReceiverBase(core)
	if err != nil {
		return nil, fmt.Errorf("failed to create baseReceiver: %w", err)
	}

	config := &core.GetConfig().Receiver
	if err := verifyWebhookConfig(config); err != nil {
		return nil, fmt.Errorf("wrong webhook config: %w", err)
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
	} // TODO: ДОБАВИТЬ ПРОВЕРКУ ЧТО ЭТО ВООБЩЕ НАШ АЙПИ

	// TODO: ДОБАВИТЬ ПРОВЕРКУ RangeIP

	// TODO: ДОБАВИТЬ ПРОВЕРКУ Port

	// TODO: ДОБАВИТЬ ПРОВЕРКУ CertFile

	// TODO: ДОБАВИТЬ ПРОВЕРКУ KeyFile

	// TODO: ДОБАВИТЬ ПРОВЕРКУ ReadTimeout

	return nil
}

func (w *Webhook) Start() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.updateHandler == nil {
		return fmt.Errorf("update handler cannot be nil")
	}

	if err := w.setupReceiverBase(); err != nil {
		return fmt.Errorf("failed to setup receiver base: %w", err)
	}

	if err := w.generateWebhookToken(); err != nil {
		return fmt.Errorf("failed to generate webhook token: %w", err)
	}

	if err := w.setupServer(); err != nil {
		return fmt.Errorf("failed to setup webhook server: %w", err)
	}

	if err := w.setupTelegramWebhook(); err != nil {
		return fmt.Errorf("failed to setup telegram webhook: %w", err)
	}

	go w.runServer()

	go w.processUpdatesFromBuffer()

	config := w.core.GetConfig().Receiver.Webhook

	w.core.GetLogger().Info("Webhook server started on %s:%s", config.ServerIP, config.Port)

	return nil
}

func (w *Webhook) generateWebhookToken() error {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Errorf("failed to generate random bytes: %w", err)
	}

	w.webhookToken = base64.URLEncoding.EncodeToString(bytes)

	return nil
}

func (w *Webhook) setupServer() error {
	mux := http.NewServeMux()
	webhookPath := "/webhook/" + w.webhookToken
	mux.HandleFunc(webhookPath, w.webhookHandler)

	config := w.core.GetConfig().Receiver.Webhook

	w.server = &http.Server{
		Addr:         config.ListenIP + ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return nil
}

func (w *Webhook) setupTelegramWebhook() error {
	config := w.core.GetConfig().Receiver.Webhook

	webhookURL := "https://" + config.ServerIP + ":" + config.Port + "/webhook/" + w.webhookToken
	webhookConfig, err := tgbotapi.NewWebhookWithCert(webhookURL, tgbotapi.FilePath(config.CertFile))
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	webhookConfig.AllowedUpdates = w.core.GetConfig().Receiver.AllowedUpdates

	_, err = w.core.GetBotAPI().Request(webhookConfig)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	w.core.GetLogger().Info("Telegram webhook configured to: %s", webhookURL)
	return nil
}

func (w *Webhook) runServer() {
	serverErr := make(chan error, 1)

	config := w.core.GetConfig().Receiver.Webhook

	go func() {
		err := w.server.ListenAndServeTLS(config.CertFile, config.KeyFile)
		serverErr <- err
	}()

	select {
	case <-w.ctx.Done():
		w.shutdownServer()
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			w.core.GetLogger().Warn("Webhook server error: %v", err)
		}
	}
}

func (w *Webhook) webhookHandler(writer http.ResponseWriter, request *http.Request) {
	log := w.core.GetLogger()

	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var update tgbotapi.Update
	if err := json.NewDecoder(request.Body).Decode(&update); err != nil {
		log.Warn("Error decoding update: %v", err)
		http.Error(writer, "Bad request", http.StatusBadRequest)
		return
	}

	if err := w.sendUpdateToBuffer(update); err != nil {
		log.Warn("Send update failed")
	}

	writer.WriteHeader(http.StatusOK)
}

func (w *Webhook) shutdownServer() {
	log := w.core.GetLogger()
	if w.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := w.server.Shutdown(ctx); err != nil {
			log.Warn("Webhook server shutdown error: %v", err)
		} else {
			log.Info("Webhook server stopped gracefully")
		}
	}
}

func (w *Webhook) Stop() error {
	log := w.core.GetLogger()

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.running {
		return nil
	}

	w.shutdownServer()

	if err := w.removeTelegramWebhook(); err != nil {
		log.Warn("Warning: failed to remove telegram webhook: %v", err)
	}

	if err := w.stopReceiverBase(); err != nil {
		return fmt.Errorf("failed to stop receiver base: %w", err)
	}

	log.Info("Webhook stopped gracefully")

	return nil
}

func (w *Webhook) removeTelegramWebhook() error {
	_, err := w.core.GetBotAPI().Request(tgbotapi.DeleteWebhookConfig{
		DropPendingUpdates: false,
	})
	return err
}

func (w *Webhook) GetType() string {
	return "webhook"
}
