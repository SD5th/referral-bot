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

func main() {
	log.Println("Запуск echo-бота...")

	// Загружаем конфигурацию
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создаем бота
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	bot.Debug = true // Включаем отладку
	log.Printf("Авторизован как %s", bot.Self.UserName)

	// Настраиваем канал для получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Println("Бот запущен и слушает сообщения...")

	// Обрабатываем входящие сообщения
	for update := range updates {
		// Игнорируем любые сообщения, кроме текстовых
		if update.Message == nil {
			continue
		}

		// Логируем полученное сообщение
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// Создаем ответное сообщение (эхо)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		// Отправляем сообщение
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
	}
}
