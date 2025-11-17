package main

import (
	"log"
	"os"
	"os/signal"
	"referral-bot/internal/bot"
	"referral-bot/internal/config"
	"referral-bot/internal/types"
	"syscall"
)

func main() {
	var err error

	log.Println("Чтение конфига...")

	var conf *config.Config
	conf, err = config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка чтения конфига: %v", err)
	}

	log.Println("Создание бота...")

	var mainBot types.BotContext
	mainBot, err = bot.NewBot(&conf.Bot)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	log.Printf("Авторизован как %s", mainBot.GetAPI().Self.UserName)

	log.Println("Бот запущен и слушает сообщения...")

	mainBot.StartReceiver()

	waitForShutdown(mainBot)
}

func waitForShutdown(bot types.BotContext) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	sig := <-sigChan
	log.Printf("Получен сигнал: %v", sig)

	if err := bot.StopReceiver(); err != nil {
		log.Printf("Ошибка при остановке: %v", err)
		os.Exit(1)
	}

	log.Println("Бот остановлен")
}
