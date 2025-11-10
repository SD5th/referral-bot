package bot

import (
	"referral-bot/internal/handlers"
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleUpdate(bot types.BotContext, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handlers.HandleMessage(bot, update.Message)
	case update.EditedMessage != nil:
		handlers.HandleEditedMessage(bot, update.EditedMessage)
	case update.ChannelPost != nil:
		handlers.HandleChannelPost(bot, update.ChannelPost)
	case update.EditedChannelPost != nil:
		handlers.HandleEditedChannelPost(bot, update.EditedChannelPost)
	case update.InlineQuery != nil:
		handlers.HandleInlineQuery(bot, update.InlineQuery)
	case update.ChosenInlineResult != nil:
		handlers.HandleChosenInlineResult(bot, update.ChosenInlineResult)
	case update.CallbackQuery != nil:
		handlers.HandleCallbackQuery(bot, update.CallbackQuery)
	case update.ShippingQuery != nil:
		handlers.HandleShippingQuery(bot, update.ShippingQuery)
	case update.PreCheckoutQuery != nil:
		handlers.HandlePreCheckoutQuery(bot, update.PreCheckoutQuery)
	case update.Poll != nil:
		handlers.HandlePoll(bot, update.Poll)
	case update.PollAnswer != nil:
		handlers.HandlePollAnswer(bot, update.PollAnswer)
	case update.MyChatMember != nil:
		handlers.HandleMyChatMember(bot, update.MyChatMember)
	case update.ChatMember != nil:
		handlers.HandleChatMember(bot, update.ChatMember)
	case update.ChatJoinRequest != nil:
		handlers.HandleChatJoinRequest(bot, update.ChatJoinRequest)
	}
}
