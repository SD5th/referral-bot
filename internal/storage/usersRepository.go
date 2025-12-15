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
		SELECT 
			id, 
			telegram_id, first_name, last_name, username, member_status, 
			invited_by_user_id, invited_by_link_id, 
			invite_link_id, 
			created_at, updated_at
		FROM users 
		WHERE 
			id = ?
		`

	var foundUser types.User
	err := r.db.SqlDB.QueryRow(query, id).Scan(
		&foundUser.ID,
		&foundUser.TelegramID,
		&foundUser.FirstName,
		&foundUser.LastName,
		&foundUser.Username,
		&foundUser.MemberStatus,
		&foundUser.InvitedByUserID,
		&foundUser.InvitedByLinkID,
		&foundUser.InviteLinkID,
		&foundUser.CreatedAt,
		&foundUser.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &foundUser, nil
}

func (r *UserRepository) GetByTelegramID(telegramID int64) (*types.User, error) {
	query := `
		SELECT 
			id, 
			telegram_id, first_name, last_name, username, member_status, 
			invited_by_user_id, invited_by_link_id, 
			invite_link_id, 
			created_at, updated_at
		FROM users 
		WHERE 
			telegram_id = ?
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
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByMinJoins(minJoins int) ([]*types.User, error) {
	query := `
		SELECT 
			id, 
			telegram_id, first_name, last_name, username, member_status, 
			invited_by_user_id, invited_by_link_id, 
			invite_link_id, 
			created_at, updated_at
		FROM users
		WHERE id IN (
			SELECT DISTINCT requester_id 
			FROM invite_links 
			WHERE unique_joins >= ?
		)
  `

	rows, err := r.db.SqlDB.Query(query, minJoins)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*types.User
	for rows.Next() {
		var user types.User

		err := rows.Scan(
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

		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) Insert(user *types.User) (*types.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	query := `
		INSERT INTO users 
			(telegram_id, first_name, last_name, username, member_status, 
			invited_by_user_id, invited_by_link_id, 
			invite_link_id,
			created_at, updated_at)
		VALUES (
			?, ?, ?, ?, ?, 
			?, ?, 
			?,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING 
			id, 
			telegram_id, first_name, last_name, username, member_status, 
			invited_by_user_id, invited_by_link_id, 
			invite_link_id, 
			created_at, updated_at
	`

	var insertedUser types.User
	err := r.db.SqlDB.QueryRow(query,
		user.TelegramID, user.FirstName, user.LastName, user.Username, user.MemberStatus,
		user.InvitedByUserID, user.InvitedByLinkID,
		user.InviteLinkID,
	).Scan(
		&insertedUser.ID,
		&insertedUser.TelegramID, &insertedUser.FirstName, &insertedUser.LastName, &insertedUser.Username, &insertedUser.MemberStatus,
		&insertedUser.InvitedByUserID, &insertedUser.InvitedByLinkID,
		&insertedUser.InviteLinkID,
		&insertedUser.CreatedAt, &insertedUser.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &insertedUser, nil
}

func (r *UserRepository) UpdateBasedOnTelegramID(user *types.User) (*types.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	query := `
		UPDATE users 
		SET 
			first_name = ?, last_name = ?, username = ?, member_status = ?, 
			invited_by_user_id = ?, invited_by_link_id = ?, 
			invite_link_id = ?, 
			updated_at = CURRENT_TIMESTAMP
		WHERE 
			telegram_id = ?
		RETURNING 
			id, 
			telegram_id, first_name, last_name, username, member_status, 
			invited_by_user_id, invited_by_link_id, 
			invite_link_id, 
			created_at, updated_at
	`

	var updatedUser types.User
	err := r.db.SqlDB.QueryRow(query,
		user.FirstName, user.LastName, user.Username, user.MemberStatus,
		user.InvitedByUserID, user.InvitedByLinkID,
		user.InviteLinkID,

		user.TelegramID,
	).Scan(
		&updatedUser.ID,
		&updatedUser.TelegramID, &updatedUser.FirstName, &updatedUser.LastName, &updatedUser.Username, &updatedUser.MemberStatus,
		&updatedUser.InvitedByUserID, &updatedUser.InvitedByLinkID,
		&updatedUser.InviteLinkID,
		&updatedUser.CreatedAt, &updatedUser.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}

func (r *UserRepository) UpsertBasedOnTelegramID(user *types.User) (*types.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	updatedUser, err := r.UpdateBasedOnTelegramID(user)
	if err == nil {
		return updatedUser, nil
	}

	if err == sql.ErrNoRows {
		return r.Insert(user)
	}

	return nil, err
}
