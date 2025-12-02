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
	SELECT id, requester_id, invite_link, name, unique_joins, created_at, updated_at
		FROM invite_links WHERE id = ?
		`

	inviteLink := new(types.InviteLink)
	err := r.db.SqlDB.QueryRow(query, id).Scan(
		&inviteLink.ID,
		&inviteLink.RequesterID,
		&inviteLink.InviteLink,
		&inviteLink.Name,
		&inviteLink.UniqueJoins,
		&inviteLink.CreatedAt,
		&inviteLink.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invite link by id (%d)", id)
	}

	return inviteLink, nil
}

func (r *InviteLinkRepository) GetByName(name string) (*types.InviteLink, error) {
	query := `
	SELECT id, requester_id, invite_link, name, unique_joins, created_at, updated_at
		FROM invite_links WHERE name = ?
		`

	inviteLink := new(types.InviteLink)
	err := r.db.SqlDB.QueryRow(query, name).Scan(
		&inviteLink.ID,
		&inviteLink.RequesterID,
		&inviteLink.InviteLink,
		&inviteLink.Name,
		&inviteLink.UniqueJoins,
		&inviteLink.CreatedAt,
		&inviteLink.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invite link by name (%s)", name)
	}

	return inviteLink, nil
}

func (r *InviteLinkRepository) Insert(inviteLink *types.InviteLink) (*types.InviteLink, error) {
	if inviteLink == nil {
		return nil, fmt.Errorf("inviteLink is nil")
	}

	query := `
	INSERT INTO invite_links 
	(requester_id, invite_link, name, unique_joins, created_at, updated_at)
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := r.db.SqlDB.Exec(
		query,
		inviteLink.RequesterID,
		inviteLink.InviteLink,
		inviteLink.Name,
		inviteLink.UniqueJoins,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert invite link: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return r.GetByID(id)
}

func (r *InviteLinkRepository) UpdateByID(inviteLink *types.InviteLink) (*types.InviteLink, error) {
	if inviteLink == nil {
		return nil, fmt.Errorf("inviteLink is nil")
	}

	dbInviteLink, err := r.GetByID(inviteLink.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite link by ID: %w", err)
	}

	if dbInviteLink == nil {
		return nil, fmt.Errorf("invite link with id: %d doesn's exist", inviteLink.ID)
	}

	query := `
		UPDATE invite_links 
		SET requester_id = ?, invite_link = ?, name = ?, unique_joins = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.SqlDB.Exec(
		query,
		inviteLink.RequesterID,
		inviteLink.InviteLink,
		inviteLink.Name,
		inviteLink.UniqueJoins,
		dbInviteLink.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update invite link: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("invite link with ID %d not found", dbInviteLink.ID)
	}

	return r.GetByID(dbInviteLink.ID)
}
