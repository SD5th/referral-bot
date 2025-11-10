package handlers

import (
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleChatMember(bot types.BotContext, chatMember *tgbotapi.ChatMemberUpdated) {}
