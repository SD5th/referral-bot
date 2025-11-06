package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Config - структура для хранения конфигурации
type Config struct {
	Token string `json:"token"`
	Debug bool   `json:"debug"`
}

// loadConfig загружает конфигурацию из config.json
func loadConfig() (*Config, error) {
	file, err := os.Open("config.json")
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия config.json: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения config.json: %v", err)
	}

	if config.Token == "" {
		return nil, fmt.Errorf("токен не указан в config.json")
	}

	return &config, nil
}

func Must[T any](v T, err error) T {
	if err != nil {
		log.Fatalf("Ошибка: %v", err)
	}
	return v
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Логируем полученное сообщение
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	// Создаем ответное сообщение (эхо)
	msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)

	// Отправляем сообщение
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

func handleEditedMessage(bot *tgbotapi.BotAPI, editedMessage *tgbotapi.Message)                {}
func handleChannelPost(bot *tgbotapi.BotAPI, channelPost *tgbotapi.Message)                    {}
func handleEditedChannelPost(bot *tgbotapi.BotAPI, editedChannelPost *tgbotapi.Message)        {}
func handleInlineQuery(bot *tgbotapi.BotAPI, inlineQuery *tgbotapi.InlineQuery)                {}
func handleChosenInlineResult(bot *tgbotapi.BotAPI, inlineResult *tgbotapi.ChosenInlineResult) {}
func handleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery)          {}
func handleShippingQuery(bot *tgbotapi.BotAPI, shippingQuery *tgbotapi.ShippingQuery)          {}
func handlePreCheckoutQuery(bot *tgbotapi.BotAPI, preCheckoutQuery *tgbotapi.PreCheckoutQuery) {}
func handlePoll(bot *tgbotapi.BotAPI, poll *tgbotapi.Poll)                                     {}
func handlePollAnswer(bot *tgbotapi.BotAPI, pollAnswer *tgbotapi.PollAnswer)                   {}
func handleMyChatMember(bot *tgbotapi.BotAPI, myChatMember *tgbotapi.ChatMemberUpdated)        {}
func handleChatMember(bot *tgbotapi.BotAPI, chatMember *tgbotapi.ChatMemberUpdated)            {}
func handleChatJoinRequest(bot *tgbotapi.BotAPI, chatJoinRequest *tgbotapi.ChatJoinRequest)    {}

func handleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handleMessage(bot, update.Message)
	case update.EditedMessage != nil:
		handleEditedMessage(bot, update.EditedMessage)
	case update.ChannelPost != nil:
		handleChannelPost(bot, update.ChannelPost)
	case update.EditedChannelPost != nil:
		handleEditedChannelPost(bot, update.EditedChannelPost)
	case update.InlineQuery != nil:
		handleInlineQuery(bot, update.InlineQuery)
	case update.ChosenInlineResult != nil:
		handleChosenInlineResult(bot, update.ChosenInlineResult)
	case update.CallbackQuery != nil:
		handleCallbackQuery(bot, update.CallbackQuery)
	case update.ShippingQuery != nil:
		handleShippingQuery(bot, update.ShippingQuery)
	case update.PreCheckoutQuery != nil:
		handlePreCheckoutQuery(bot, update.PreCheckoutQuery)
	case update.Poll != nil:
		handlePoll(bot, update.Poll)
	case update.PollAnswer != nil:
		handlePollAnswer(bot, update.PollAnswer)
	case update.MyChatMember != nil:
		handleMyChatMember(bot, update.MyChatMember)
	case update.ChatMember != nil:
		handleChatMember(bot, update.ChatMember)
	case update.ChatJoinRequest != nil:
		handleChatJoinRequest(bot, update.ChatJoinRequest)
	}

}

func main() {
	log.Println("Запуск бота...")

	// Загружаем конфигурацию
	config := Must(loadConfig())

	// Создаем бота
	bot := Must(tgbotapi.NewBotAPI(config.Token))
	bot.Debug = config.Debug
	log.Printf("Авторизован как %s", bot.Self.UserName)

	// Настраиваем канал для получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Println("Бот запущен и слушает сообщения...")

	// Обрабатываем входящие сообщения
	for update := range updates {
		handleUpdate(bot, update)
	}
}
