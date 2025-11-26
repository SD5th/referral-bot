package main

import (
	"fmt"
	"os"
	"os/signal"
	"referral-bot/internal/bot"
	"referral-bot/internal/config"
	"referral-bot/internal/core"
	"referral-bot/internal/handlers"
	"referral-bot/internal/interfaces"
	"referral-bot/internal/logger"
	"referral-bot/internal/receivers"
	"syscall"
	"time"
)

func main() {
	core := core.NewCore()

	logger := logger.NewStdLogger()

	core.SetLogger(logger)
	log := core.GetLogger()

	startTime := time.Now()

	log.Info("Starting application...")

	log.Info("Loading configuration...")
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}
	core.SetConfig(conf)
	config := core.GetConfig()
	log.Info("Configuration loaded")

	fmt.Printf("%+v\n", config)

	/*
		mainDB, err := storage.NewDatabase()
		if err != nil {
			log.Fatal("Failed to initialize database: %v", err)
			}
	*/

	log.Info("Initializing bot...")
	botAPI, err := bot.SetupBotAPI(&config.BotAPI)
	if err != nil {
		log.Fatal("Failed to create bot: %v", err)
	}
	core.SetBotAPI(botAPI)

	botUser := core.GetBotAPI().Self
	log.Info("Bot authorized: @%s (ID: %d)", botUser.UserName, botUser.ID)

	updateReceiver, err := receivers.NewUpdateReceiver(core)
	if err != nil {
		log.Fatal("Failed to create update receiver: %v", err)
	}

	updateHandler, err := handlers.NewUpdateHandler(core, updateReceiver)
	if err != nil {
		log.Fatal("Failed to create update handler: %v", err)
	}

	if err := updateReceiver.SetUpdateHandler(updateHandler); err != nil {
		log.Fatal("Failed to set update handler: %v", err)
	}

	log.Info("Starting update receiver...")
	if err := updateReceiver.Start(); err != nil {
		log.Fatal("Failed to start update receiver")
	}

	log.Info("Receiver started")

	startupTime := time.Since(startTime)
	log.Info("Application started successfully (took %v)", startupTime.Round(time.Millisecond))

	waitForShutdown(core, updateReceiver)
}

func waitForShutdown(core *core.Core, updateReceiver interfaces.UpdateReceiverInterface) {
	log := core.GetLogger()

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

	if err := updateReceiver.Stop(); err != nil {
		log.Error("Error stopping receiver: %v", err)
		os.Exit(1)
	}

	log.Info("Shutdown completed")
}
