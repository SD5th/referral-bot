package main

import (
	"os"
	"os/signal"
	"referral-bot/internal/bot"
	"referral-bot/internal/config"
	"referral-bot/internal/core"
	"referral-bot/internal/handlers"
	"referral-bot/internal/interfaces"
	"referral-bot/internal/logger"
	"referral-bot/internal/receivers"
	"referral-bot/internal/services"
	"referral-bot/internal/storage"
	"syscall"
)

func main() {
	core := core.NewCore()

	logger := logger.NewStdLogger()

	core.SetLogger(logger)
	log := core.GetLogger()

	log.Info("Starting application...")

	log.Info("Loading configuration...")
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}
	core.SetConfig(conf)
	config := core.GetConfig()
	log.Info("Configuration loaded")

	//fmt.Printf("%+v\n", config)

	db, err := storage.NewDatabase(core, "./data/bot.db")
	if err != nil {
		log.Fatal("Failed to initialize database: %v", err)
	}

	// repositories
	activeChannelRepository := storage.NewActiveChannelRepository(core, db)
	channelActivityRepository := storage.NewChannelActivityRepository(core, db)
	userRepository := storage.NewUserRepository(core, db)
	adminRepository := storage.NewAdminRepository(core, db)
	inviteLinkRepository := storage.NewInviteLinkRepository(core, db)

	// services
	activeChannelService := services.NewActiveChannelService(core, activeChannelRepository, userRepository, inviteLinkRepository, channelActivityRepository)
	inviteLinkService := services.NewInviteLinkService(core, activeChannelRepository, userRepository, inviteLinkRepository, channelActivityRepository)
	userService := services.NewUserService(core, activeChannelRepository, userRepository, inviteLinkRepository, channelActivityRepository)
	adminService := services.NewAdminService(core, activeChannelRepository, userRepository, adminRepository, inviteLinkRepository, channelActivityRepository)
	tgUtilsService := services.NewTGUtilsService(core)
	channelActivityService := services.NewChannelActivityService(core, activeChannelRepository, userRepository, adminRepository, inviteLinkRepository, channelActivityRepository)

	updateHandler, err := handlers.NewUpdateHandler(core, activeChannelService, userService, adminService, tgUtilsService, inviteLinkService, channelActivityService)
	if err != nil {
		log.Fatal("Failed to create update handler: %v", err)
	}

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

	if err := updateReceiver.SetUpdateHandler(updateHandler); err != nil {
		log.Fatal("Failed to set update handler: %v", err)
	}

	log.Info("Starting update receiver...")
	if err := updateReceiver.Start(); err != nil {
		log.Fatal("Failed to start update receiver")
	}
	log.Info("Receiver started")

	waitForShutdown(core, updateReceiver)
}

func waitForShutdown(core *core.Core, updateReceiver interfaces.UpdateReceiver) {
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
