package main

import (
	"log"
	"os"
	"os/signal"
	"referral-bot/internal/bot"
	"syscall"
)

func main() {
	log.Println("Создание бота...")

	mainBot, err := bot.NewBot()
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	log.Printf("Авторизован как %s", mainBot.API.Self.UserName)

	log.Println("Бот запущен и слушает сообщения...")

	(*mainBot.UpdateReceiver).Start()

	waitForShutdown(mainBot)
}

func waitForShutdown(bot *bot.Bot) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	sig := <-sigChan
	log.Printf("Получен сигнал: %v", sig)

	if err := (*bot.UpdateReceiver).Stop(); err != nil {
		log.Printf("Ошибка при остановке: %v", err)
		os.Exit(1)
	}

	log.Println("Бот остановлен")
}
