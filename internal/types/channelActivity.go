package types

import "time"

type ChannelActivity struct {
	ID int64 `json:"id"`

	ChannelTelegramID int64 `json:"channel_telegram_id"`

	UserTelegramID int64  `json:"user_telegram_id"`
	UserFirstName  string `json:"user_first_name"`
	UserLastName   string `json:"user_last_name,omitempty"`
	UserUsername   string `json:"user_username,omitempty"`

	InviterTelegramID *int64 `json:"inviter_telegram_id,omitempty"`
	InviterFirstName  string `json:"inviter_first_name,omitempty"`
	InviterLastName   string `json:"inviter_last_name,omitempty"`
	InviterUsername   string `json:"inviter_username,omitempty"`

	InviteLinkURL  string `json:"invite_link_url,omitempty"`
	InviteLinkName string `json:"invite_link_name,omitempty"`

	OldStatus MemberStatus `json:"old_status"`
	NewStatus MemberStatus `json:"new_status"`

	CreatedAt time.Time `json:"created_at"`
}
