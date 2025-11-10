package handlers

import (
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandlePreCheckoutQuery(bot types.BotContext, preCheckoutQuery *tgbotapi.PreCheckoutQuery) {}
