package handlers

import (
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMyChatMember(bot types.BotContext, myChatMember *tgbotapi.ChatMemberUpdated) {}
