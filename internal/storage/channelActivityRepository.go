package storage

import (
	"database/sql"
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/types"
)

type ChannelActivityRepository struct {
	core *core.Core
	db   *Database
}

func NewChannelActivityRepository(core *core.Core, db *Database) *ChannelActivityRepository {
	return &ChannelActivityRepository{
		core: core,
		db:   db,
	}
}

func (r *ChannelActivityRepository) GetByID(id int64) (*types.ChannelActivity, error) {
	query := `
		SELECT 
			id, 
			channel_telegram_id, user_telegram_id, user_first_name, user_last_name, user_username, 
			inviter_telegram_id, inviter_first_name, inviter_last_name, inviter_username, 
			invite_link_url, invite_link_name, 
			old_status, new_status, 
			created_at
		FROM channel_activity 
		WHERE 
			id = ?
	`

	activity := new(types.ChannelActivity)
	var inviterTelegramID sql.NullInt64

	err := r.db.SqlDB.QueryRow(query, id).Scan(
		&activity.ID,
		&activity.ChannelTelegramID, &activity.UserTelegramID, &activity.UserFirstName, &activity.UserLastName, &activity.UserUsername,
		&inviterTelegramID, &activity.InviterFirstName, &activity.InviterLastName, &activity.InviterUsername,
		&activity.InviteLinkURL, &activity.InviteLinkName,
		&activity.OldStatus, &activity.NewStatus,
		&activity.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel activity by id (%d)", id)
	}

	activity.InviterTelegramID = nullInt64ToPtr(inviterTelegramID)
	return activity, nil
}

func (r *ChannelActivityRepository) Insert(activity *types.ChannelActivity) (*types.ChannelActivity, error) {
	if activity == nil {
		return nil, fmt.Errorf("channelActivity is nil")
	}

	query := `
		INSERT INTO channel_activity (
			channel_telegram_id, 
			user_telegram_id, user_first_name, user_last_name, user_username, 
			inviter_telegram_id, inviter_first_name, inviter_last_name, inviter_username, 
			invite_link_url, invite_link_name, 
			old_status, new_status, 
			created_at
		)
		VALUES (
			?, 
			?, ?, ?, ?, 
			?, ?, ?, ?, 
			?, ?, 
			?, ?, 
			CURRENT_TIMESTAMP
		)
		RETURNING 
			id, 
			channel_telegram_id, 
			user_telegram_id, user_first_name, user_last_name, user_username, 
			inviter_telegram_id, inviter_first_name, inviter_last_name, inviter_username, 
			invite_link_url, invite_link_name, 
			old_status, new_status, 
			created_at
	`

	var insertedActivity types.ChannelActivity
	//var inviterTelegramID sql.NullInt64

	err := r.db.SqlDB.QueryRow(query,
		activity.ChannelTelegramID,
		activity.UserTelegramID, activity.UserFirstName, activity.UserLastName, activity.UserUsername,
		nullInt64(activity.InviterTelegramID), activity.InviterFirstName, activity.InviterLastName, activity.InviterUsername,
		activity.InviteLinkURL, activity.InviteLinkName,
		activity.OldStatus, activity.NewStatus,
	).Scan(
		&insertedActivity.ID,
		&insertedActivity.ChannelTelegramID, &insertedActivity.UserTelegramID, &insertedActivity.UserFirstName, &insertedActivity.UserLastName, &insertedActivity.UserUsername,
		&insertedActivity.InviterTelegramID, &insertedActivity.InviterFirstName, &insertedActivity.InviterLastName, &insertedActivity.InviterUsername,
		&insertedActivity.InviteLinkURL, &insertedActivity.InviteLinkName,
		&insertedActivity.OldStatus, &insertedActivity.NewStatus,
		&insertedActivity.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	//insertedActivity.InviterTelegramID = nullInt64ToPtr(inviterTelegramID)

	return &insertedActivity, nil
}

func nullInt64(ptr *int64) interface{} {
	if ptr == nil {
		return nil
	}
	return *ptr
}

func nullInt64ToPtr(nullInt sql.NullInt64) *int64 {
	if !nullInt.Valid {
		return nil
	}
	ptr := new(int64)
	*ptr = nullInt.Int64
	return ptr
}
