package handlers

import (
	"referral-bot/internal/core"
	"referral-bot/internal/interfaces"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UpdateHandler struct {
	core           *core.Core
	updateReceiver interfaces.UpdateReceiverInterface
}

func NewUpdateHandler(core *core.Core, updateReceiver interfaces.UpdateReceiverInterface) (*UpdateHandler, error) {
	return &UpdateHandler{
		core:           core,
		updateReceiver: updateReceiver,
	}, nil
}

func (u *UpdateHandler) HandleUpdate(update tgbotapi.Update) error {
	switch {
	case update.Message != nil:
		u.handleMessage(update.Message)
	case update.EditedMessage != nil:
		u.handleEditedMessage(update.EditedMessage)
	case update.ChannelPost != nil:
		u.handleChannelPost(update.ChannelPost)
	case update.EditedChannelPost != nil:
		u.handleEditedChannelPost(update.EditedChannelPost)
	case update.InlineQuery != nil:
		u.handleInlineQuery(update.InlineQuery)
	case update.ChosenInlineResult != nil:
		u.handleChosenInlineResult(update.ChosenInlineResult)
	case update.CallbackQuery != nil:
		u.handleCallbackQuery(update.CallbackQuery)
	case update.ShippingQuery != nil:
		u.handleShippingQuery(update.ShippingQuery)
	case update.PreCheckoutQuery != nil:
		u.handlePreCheckoutQuery(update.PreCheckoutQuery)
	case update.Poll != nil:
		u.handlePoll(update.Poll)
	case update.PollAnswer != nil:
		u.handlePollAnswer(update.PollAnswer)
	case update.MyChatMember != nil:
		u.handleMyChatMember(update.MyChatMember)
	case update.ChatMember != nil:
		u.handleChatMember(update.ChatMember)
	case update.ChatJoinRequest != nil:
		u.handleChatJoinRequest(update.ChatJoinRequest)
	}
	return nil
}
