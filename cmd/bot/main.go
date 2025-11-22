package main

import (
	"os"
	"os/signal"
	"referral-bot/internal/bot"
	"referral-bot/internal/config"
	"referral-bot/internal/logger"
	"referral-bot/internal/types"
	"syscall"
	"time"
)

func main() {
	var log types.LoggerContext
	log = logger.NewStdLogger()
	startTime := time.Now()

	log.Info("Starting application...")

	log.Info("Loading configuration...")
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}
	log.Info("Configuration loaded")

	log.Info("Initializing bot...")
	var mainBot types.BotContext
	mainBot, err = bot.NewBot(&conf.Bot, log)
	if err != nil {
		log.Fatal("Failed to create bot: %v", err)
	}

	log = mainBot.GetLogger()

	botUser := mainBot.GetAPI().Self
	log.Info("Bot authorized: @%s (ID: %d)", botUser.UserName, botUser.ID)

	log.Info("Starting update receiver...")
	mainBot.StartReceiver()
	log.Info("Receiver started")

	startupTime := time.Since(startTime)
	log.Info("Application started successfully (took %v)", startupTime.Round(time.Millisecond))

	waitForShutdown(mainBot)
}

func waitForShutdown(bot types.BotContext) {
	log := bot.GetLogger()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	log.Info("Waiting for shutdown signal...")

	sig := <-sigChan
	log.Info("Received signal: %v", sig)
	log.Info("Initiating shutdown...")

	if err := bot.StopReceiver(); err != nil {
		log.Error("Error stopping receiver: %v", err)
		os.Exit(1)
	}

	log.Info("Shutdown completed")
}
