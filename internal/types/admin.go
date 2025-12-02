package types

import (
	"time"
)

type Admin struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name,omitempty"`
	Username   string    `json:"username,omitempty"` // without @
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
