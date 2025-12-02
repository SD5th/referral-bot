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
	SELECT id, user_id, invited_by_user_id, invited_by_link_id, old_status, new_status, created_at
		FROM channel_activity WHERE id = ?
		`

	channelActivity := new(types.ChannelActivity)
	err := r.db.SqlDB.QueryRow(query, id).Scan(
		&channelActivity.ID,
		&channelActivity.UserID,
		&channelActivity.InvitedByUserID,
		&channelActivity.InvitedByLinkID,
		&channelActivity.OldStatus,
		&channelActivity.NewStatus,
		&channelActivity.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get channel activity by id (%d)", id)
	}

	return channelActivity, nil
}

func (r *ChannelActivityRepository) Insert(channelActivity *types.ChannelActivity) (*types.ChannelActivity, error) {
	if channelActivity == nil {
		return nil, fmt.Errorf("channelActivity is nil")
	}

	query := `
	INSERT INTO channel_activity 
	(user_id, invited_by_user_id, invited_by_link_id, old_status, new_status, created_at)
	VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	result, err := r.db.SqlDB.Exec(
		query,
		channelActivity.UserID,
		channelActivity.InvitedByUserID,
		channelActivity.InvitedByLinkID,
		channelActivity.OldStatus,
		channelActivity.NewStatus,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert channel activity: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return r.GetByID(id)
}
