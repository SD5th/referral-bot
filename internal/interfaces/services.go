package interfaces

import (
	"referral-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ActiveChannelService interface {
	Register(telegramID int64) (*types.Channel, error)
	Get() (*types.Channel, error)
	IsActive() bool
	Leave() error
}

type AdminService interface {
	Register(tgUser *tgbotapi.User) (*types.Admin, error)
	IsAdmin(telegramID int64) (bool, error)
}

type UserService interface {
	UpdateOrRegister(tgUser *tgbotapi.User) (*types.User, error)

	ProcessJoin(chatMemberUpdated *tgbotapi.ChatMemberUpdated) (*types.User, error)
	ProcessLeave(chatMemberUpdated *tgbotapi.ChatMemberUpdated) (*types.User, error)

	RequestMemberStatusByTelegramID(telegramID int64) (types.MemberStatus, error)
}

type TGUtilsService interface {
	SendMessage(chatID int64, text string) error
	SendTryAgainLater(chatID int64) error
	//LeaveChannel(channelID int64) error
}

type InviteLinkService interface {
	GetByRequesterTelegramID(telegramID int64) (*types.InviteLink, error)
	GetOrCreateByRequesterTelegramID(telegramID int64) (*types.InviteLink, error)
}

type ChannelActivityService interface {
	AddFromUpdate(update *tgbotapi.ChatMemberUpdated) (*types.ChannelActivity, error)
}
