package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		handleMessage(bot, update.Message)
	case update.EditedMessage != nil:
		handleEditedMessage(bot, update.EditedMessage)
	case update.ChannelPost != nil:
		handleChannelPost(bot, update.ChannelPost)
	case update.EditedChannelPost != nil:
		handleEditedChannelPost(bot, update.EditedChannelPost)
	case update.InlineQuery != nil:
		handleInlineQuery(bot, update.InlineQuery)
	case update.ChosenInlineResult != nil:
		handleChosenInlineResult(bot, update.ChosenInlineResult)
	case update.CallbackQuery != nil:
		handleCallbackQuery(bot, update.CallbackQuery)
	case update.ShippingQuery != nil:
		handleShippingQuery(bot, update.ShippingQuery)
	case update.PreCheckoutQuery != nil:
		handlePreCheckoutQuery(bot, update.PreCheckoutQuery)
	case update.Poll != nil:
		handlePoll(bot, update.Poll)
	case update.PollAnswer != nil:
		handlePollAnswer(bot, update.PollAnswer)
	case update.MyChatMember != nil:
		handleMyChatMember(bot, update.MyChatMember)
	case update.ChatMember != nil:
		handleChatMember(bot, update.ChatMember)
	case update.ChatJoinRequest != nil:
		handleChatJoinRequest(bot, update.ChatJoinRequest)
	}
}
