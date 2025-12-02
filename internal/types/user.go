package types

import (
	"time"
)

type User struct {
	ID              int64        `json:"id"`
	TelegramID      int64        `json:"telegram_id"`
	FirstName       string       `json:"first_name"`
	LastName        string       `json:"last_name,omitempty"`
	Username        string       `json:"username,omitempty"` // without @
	MemberStatus    MemberStatus `json:"member_status"`
	InvitedByUserID *int64       `json:"invited_by_user_id,omitempty"` // can be nil
	InvitedByLinkID *int64       `json:"invited_by_link_id,omitempty"` // can be nil
	InviteLinkID    *int64       `json:"invite_link_id,omitempty"`     // user's invite link, can be nil
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

type MemberStatus string

const (
	MemberStatusCreator       MemberStatus = "creator"
	MemberStatusAdministrator MemberStatus = "administrator"
	MemberStatusMember        MemberStatus = "member"
	MemberStatusRestricted    MemberStatus = "restricted"
	MemberStatusLeft          MemberStatus = "left"
	MemberStatusKicked        MemberStatus = "kicked"
)

/*
	func UserFromTgChatMember(tgChatMember *tgbotapi.ChatMember) *User {
		if tgChatMember == nil || tgChatMember.User == nil {
			return nil
		}
		return &User{
			//ID: getting from DB
			TelegramID:      tgChatMember.User.ID,
			Username:        tgChatMember.User.UserName,
			FirstName:       tgChatMember.User.FirstName,
			LastName:        tgChatMember.User.LastName,
			Status:          UserStatus(tgChatMember.Status),
			InvitedByUserID: nil, // by default
			InvitedByLinkID: nil, // by default
			InviteLinkID:    nil, // by default
			//CreatedAt: DB debug info
			//UpdatedAt: DB debug info
		}
	}

	func UserFromTgUser(tgUser *tgbotapi.User) *User {
		if tgUser == nil {
			return nil
		}
		return &User{
			//ID: getting from DB
			TelegramID:      tgUser.ID,
			Username:        tgUser.UserName,
			FirstName:       tgUser.FirstName,
			LastName:        tgUser.LastName,
			Status:          Member, // by default
			InvitedByUserID: nil,    // by default
			InvitedByLinkID: nil,    // by default
			InviteLinkID:    nil,    // by default
			//CreatedAt: DB debug info
			//UpdatedAt: DB debug info
		}
	}
*/
func (user *User) IsMember() bool {
	if user == nil {
		return false
	}
	if user.MemberStatus == MemberStatusCreator || user.MemberStatus == MemberStatusAdministrator || user.MemberStatus == MemberStatusMember || user.MemberStatus == MemberStatusRestricted {
		return true
	}
	return false
}

func (user *User) IsAdmin() bool {
	if user == nil {
		return false
	}
	if user.MemberStatus == MemberStatusCreator || user.MemberStatus == MemberStatusAdministrator {
		return true
	}
	return false
}
