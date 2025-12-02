package types

import (
	"time"
)

type Channel struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`           // Channel id in Telegram
	Type       ChatType  `json:"type"`                  // Type of chat, can be either “private”, “group”, “supergroup” or “channel”
	Username   string    `json:"username,omitempty"`    //
	Title      string    `json:"title,omitempty"`       //
	InviteLink string    `json:"invite_link,omitempty"` //
	CreatedAt  time.Time `json:"created_at"`            //
	UpdatedAt  time.Time `json:"updated_at"`            //
}

type ChatType string

const (
	ChatTypePrivateType    ChatType = "private"
	ChatTypeGroupType      ChatType = "group"
	ChatTypeSupergroupType ChatType = "supergroup"
	ChatTypeChannel        ChatType = "channel"
)
