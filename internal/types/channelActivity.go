package types

import "time"

type ChannelActivity struct {
	ID              int64        `json:"id"`
	UserID          int64        `json:"user_id"`            // FK to users(id)
	InvitedByUserID *int64       `json:"invited_by_user_id"` // FK to users(id) - can be nil
	InvitedByLinkID *int64       `json:"invited_by_link_id"` // FK to invite_links(id) - can be nil
	OldStatus       MemberStatus `json:"old_status"`         // “creator”, “administrator”, “member”, “restricted”, “left” or “kicked”
	NewStatus       MemberStatus `json:"new_status"`         // “creator”, “administrator”, “member”, “restricted”, “left” or “kicked”
	CreatedAt       time.Time    `json:"created_at"`
}
