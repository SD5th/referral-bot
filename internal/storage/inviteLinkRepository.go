package storage

import (
	"database/sql"
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/types"
)

type InviteLinkRepository struct {
	core *core.Core
	db   *Database
}

func NewInviteLinkRepository(core *core.Core, db *Database) *InviteLinkRepository {
	return &InviteLinkRepository{
		core: core,
		db:   db,
	}
}

func (r *InviteLinkRepository) GetByID(id int64) (*types.InviteLink, error) {
	query := `
		SELECT 
			id, 
			requester_id, 
			url, name, 
			unique_joins, 
			created_at, updated_at
		FROM invite_links 
		WHERE 
			id = ?
	`

	var inviteLink types.InviteLink
	err := r.db.SqlDB.QueryRow(query, id).Scan(
		&inviteLink.ID,
		&inviteLink.RequesterID,
		&inviteLink.URL, &inviteLink.Name,
		&inviteLink.UniqueJoins,
		&inviteLink.CreatedAt, &inviteLink.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &inviteLink, nil
}

func (r *InviteLinkRepository) GetByName(name string) (*types.InviteLink, error) {
	query := `
		SELECT 
			id, 
			requester_id, 
			url, name, 
			unique_joins, 
			created_at, updated_at
		FROM 
			invite_links 
		WHERE 
			name = ?
	`

	var inviteLink types.InviteLink
	err := r.db.SqlDB.QueryRow(query, name).Scan(
		&inviteLink.ID,
		&inviteLink.RequesterID,
		&inviteLink.URL, &inviteLink.Name,
		&inviteLink.UniqueJoins,
		&inviteLink.CreatedAt, &inviteLink.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &inviteLink, nil
}

func (r *InviteLinkRepository) Insert(inviteLink *types.InviteLink) (*types.InviteLink, error) {
	if inviteLink == nil {
		return nil, fmt.Errorf("inviteLink is nil")
	}

	query := `
		INSERT INTO invite_links (
			requester_id, 
			url, name, 
			unique_joins, 
			created_at, updated_at
		)
		VALUES (
			?, 
			?, ?, 
			?, 
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING
			id, 
			requester_id, 
			url, name, 
			unique_joins, 
			created_at, updated_at
	`

	var insertedLink types.InviteLink
	err := r.db.SqlDB.QueryRow(query,
		inviteLink.RequesterID,
		inviteLink.URL, inviteLink.Name,
		inviteLink.UniqueJoins,
	).Scan(
		&insertedLink.ID,
		&insertedLink.RequesterID,
		&insertedLink.URL, &insertedLink.Name,
		&insertedLink.UniqueJoins,
		&insertedLink.CreatedAt, &insertedLink.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &insertedLink, nil
}

func (r *InviteLinkRepository) IncreaseCounterByID(id int64) (*types.InviteLink, error) {
	query := `
		UPDATE invite_links 
		SET 
			unique_joins = unique_joins + 1, 
			updated_at = CURRENT_TIMESTAMP
		WHERE 
			id = ?
		RETURNING
			id, 
			requester_id, 
			url, name, 
			unique_joins, 
			created_at, updated_at	
	`

	var updatedLink types.InviteLink
	err := r.db.SqlDB.QueryRow(query,
		id,
	).Scan(
		&updatedLink.ID,
		&updatedLink.RequesterID,
		&updatedLink.URL, &updatedLink.Name,
		&updatedLink.UniqueJoins,
		&updatedLink.CreatedAt, &updatedLink.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return &updatedLink, nil
}

func (r *InviteLinkRepository) DecreaseCounterByID(id int64) (*types.InviteLink, error) {
	query := `
		UPDATE invite_links 
		SET 
			unique_joins = unique_joins - 1, 
			updated_at = CURRENT_TIMESTAMP
		WHERE 
			id = ?
		RETURNING
			id, 
			requester_id, 
			url, name, 
			unique_joins, 
			created_at, updated_at	
	`

	var updatedLink types.InviteLink
	err := r.db.SqlDB.QueryRow(query,
		id,
	).Scan(
		&updatedLink.ID,
		&updatedLink.RequesterID,
		&updatedLink.URL, &updatedLink.Name,
		&updatedLink.UniqueJoins,
		&updatedLink.CreatedAt, &updatedLink.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return &updatedLink, nil
}
