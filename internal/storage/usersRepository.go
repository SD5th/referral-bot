package storage

import (
	"database/sql"
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/types"
)

type UserRepository struct {
	core *core.Core
	db   *Database
}

func NewUserRepository(core *core.Core, db *Database) *UserRepository {
	return &UserRepository{
		core: core,
		db:   db,
	}
}

func (r *UserRepository) GetByID(id int64) (*types.User, error) {
	query := `
	SELECT id, telegram_id, first_name, last_name, username, member_status, invited_by_user_id, invited_by_link_id, invite_link_id, created_at, updated_at
		FROM users WHERE id = ?
		`

	user := new(types.User)
	err := r.db.SqlDB.QueryRow(query, id).Scan(
		&user.ID,
		&user.TelegramID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.MemberStatus,
		&user.InvitedByUserID,
		&user.InvitedByLinkID,
		&user.InviteLinkID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id")
	}

	return user, nil
}

func (r *UserRepository) GetByTelegramID(telegramID int64) (*types.User, error) {
	query := `
	SELECT id, telegram_id, first_name, last_name, username, member_status, invited_by_user_id, invited_by_link_id, invite_link_id, created_at, updated_at
		FROM users WHERE telegram_id = ?
		`

	user := new(types.User)
	err := r.db.SqlDB.QueryRow(query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.MemberStatus,
		&user.InvitedByUserID,
		&user.InvitedByLinkID,
		&user.InviteLinkID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by telegram id")
	}

	return user, nil
}

func (r *UserRepository) Insert(user *types.User) (*types.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	query := `
	INSERT INTO users 
	(telegram_id, first_name, last_name, username, member_status, invited_by_user_id, invited_by_link_id, invite_link_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := r.db.SqlDB.Exec(
		query,
		user.TelegramID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.MemberStatus,
		user.InvitedByUserID,
		user.InvitedByLinkID,
		user.InviteLinkID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return r.GetByID(id)
}

func (r *UserRepository) UpdateBasedOnTelegramID(user *types.User) (*types.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	dbUser, err := r.GetByTelegramID(user.TelegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}

	if dbUser == nil {
		return nil, fmt.Errorf("user with telegram id: %d doesn's exist", user.TelegramID)
	}

	query := `
		UPDATE users 
		SET first_name = ?, last_name = ?, username = ?, member_status = ?, invited_by_user_id = ?, invited_by_link_id = ?, invite_link_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.SqlDB.Exec(
		query,
		user.FirstName,
		user.LastName,
		user.Username,
		user.MemberStatus,
		user.InvitedByUserID,
		user.InvitedByLinkID,
		user.InviteLinkID,
		dbUser.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("user with ID %d not found", dbUser.ID)
	}

	return r.GetByID(dbUser.ID)
}

func (r *UserRepository) UpsertBasedOnTelegramID(user *types.User) (*types.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	dbUser, err := r.GetByTelegramID(user.TelegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by telegram ID: %w", err)
	}

	if dbUser == nil {
		return r.Insert(user)
	}

	return r.UpdateBasedOnTelegramID(user)
}
