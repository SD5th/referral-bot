package interfaces

import "referral-bot/internal/types"

type ActiveChannelRepository interface {
	Set(channel *types.Channel) (*types.Channel, error)
	Get() (*types.Channel, error)
	Delete() error
	Update(channel *types.Channel) (*types.Channel, error)
}

type AdminRepository interface {
	GetByID(id int64) (*types.Admin, error)
	GetByTelegramID(telegramID int64) (*types.Admin, error)
	Insert(admin *types.Admin) (*types.Admin, error)
}

type UserRepository interface {
	GetByID(id int64) (*types.User, error)
	GetByTelegramID(telegramID int64) (*types.User, error)
	Insert(user *types.User) (*types.User, error)
	UpdateBasedOnTelegramID(user *types.User) (*types.User, error)
	UpsertBasedOnTelegramID(user *types.User) (*types.User, error)
}

type InviteLinkRepository interface {
	GetByID(id int64) (*types.InviteLink, error)
	GetByName(name string) (*types.InviteLink, error)
	Insert(inviteLink *types.InviteLink) (*types.InviteLink, error)
	UpdateByID(inviteLink *types.InviteLink) (*types.InviteLink, error)
	//Upsert(inviteLink *types.InviteLink) (*types.InviteLink, error)
}

type ChannelActivityRepository interface {
	GetByID(id int64) (*types.ChannelActivity, error)
	Insert(channelActivity *types.ChannelActivity) (*types.ChannelActivity, error)
}
