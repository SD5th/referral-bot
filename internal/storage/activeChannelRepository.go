package storage

import (
	"database/sql"
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/types"
)

type ActiveChannelRepository struct {
	core *core.Core
	db   *Database
}

func NewActiveChannelRepository(core *core.Core, db *Database) *ActiveChannelRepository {
	return &ActiveChannelRepository{
		core: core,
		db:   db,
	}
}

// Create создает новый канал
func (r *ActiveChannelRepository) Set(channel *types.Channel) (*types.Channel, error) {
	if channel == nil {
		return nil, fmt.Errorf("channel cannot be nil")
	}

	query := `
		INSERT INTO active_channel (
			telegram_id, type, username, title, 
			invite_link, 
			created_at, updated_at
		)
		VALUES (
			?, ?, ?, ?, 
			?, 
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING
			id, 
			telegram_id, type, username, title, 
			invite_link, 
			created_at, updated_at
	`

	var insertedChannel types.Channel
	err := r.db.SqlDB.QueryRow(query,
		channel.TelegramID, channel.Type, channel.Username, channel.Title,
		channel.InviteLink,
	).Scan(
		&insertedChannel.ID,
		&insertedChannel.TelegramID, &insertedChannel.Type, &insertedChannel.Username, &insertedChannel.Title,
		&insertedChannel.InviteLink,
		&insertedChannel.CreatedAt, &insertedChannel.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &insertedChannel, nil
}

func (r *ActiveChannelRepository) Get() (*types.Channel, error) {
	query := `
		SELECT 
			id, 
			telegram_id, type, username, title, 
			invite_link, 
			created_at, updated_at
		FROM active_channel 
		WHERE 
			id = 1
	`

	var channel types.Channel
	err := r.db.SqlDB.QueryRow(query).Scan(
		&channel.ID,
		&channel.TelegramID, &channel.Type, &channel.Username, &channel.Title,
		&channel.InviteLink,
		&channel.CreatedAt, &channel.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &channel, nil
}

func (r *ActiveChannelRepository) Delete() error {
	query := `
		DELETE
		FROM active_channel 
		WHERE 
			id = 1
	`
	result, err := r.db.SqlDB.Exec(query)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("nothing happened")
	}

	return nil
}

func (r *ActiveChannelRepository) Update(channel *types.Channel) (*types.Channel, error) {
	if channel == nil {
		return nil, fmt.Errorf("channel cannot be nil")
	}

	query := `
		UPDATE active_channel 
		SET 
			type = ?, username = ?, title = ?, 
			invite_link = ?, 
			updated_at = CURRENT_TIMESTAMP
		WHERE 
			id = 1
		RETURNING
			id, 
			telegram_id, type, username, title, 
			invite_link, 
			created_at, updated_at
	`

	var updatedChannel types.Channel
	err := r.db.SqlDB.QueryRow(query,
		channel.TelegramID, channel.Type, channel.Username, channel.Title,
		channel.InviteLink,
	).Scan(
		&updatedChannel.ID,
		&updatedChannel.TelegramID, &updatedChannel.Type, &updatedChannel.Username, &updatedChannel.Title,
		&updatedChannel.InviteLink,
		&updatedChannel.CreatedAt, &updatedChannel.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update channel: %w", err)
	}

	return &updatedChannel, nil
}
