package types

import "time"

type InviteLink struct {
	ID          int64     `json:"id"`
	RequesterID int64     `json:"requester_id"` // FK to users(id) - who requested bot to create link
	InviteLink  string    `json:"invite_link"`  // Invite link
	Name        string    `json:"name"`         // Invite link name
	UniqueJoins int       `json:"unique_joins"` // Currently joined unique users
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
