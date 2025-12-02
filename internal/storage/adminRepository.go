package storage

import (
	"database/sql"
	"fmt"
	"referral-bot/internal/core"
	"referral-bot/internal/types"
)

type AdminRepository struct {
	core *core.Core
	db   *Database
}

func NewAdminRepository(core *core.Core, db *Database) *AdminRepository {
	return &AdminRepository{
		core: core,
		db:   db,
	}
}

func (r *AdminRepository) GetByID(id int64) (*types.Admin, error) {
	query := `
	SELECT id, telegram_id, first_name, last_name, username, created_at, updated_at
		FROM admins WHERE id = ?
		`

	admin := new(types.Admin)
	err := r.db.SqlDB.QueryRow(query, id).Scan(
		&admin.ID,
		&admin.TelegramID,
		&admin.FirstName,
		&admin.LastName,
		&admin.Username,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admin by id")
	}

	return admin, nil
}

func (r *AdminRepository) GetByTelegramID(telegramID int64) (*types.Admin, error) {
	query := `
	SELECT id, telegram_id, first_name, last_name, username, created_at, updated_at
		FROM admins WHERE telegram_id = ?
		`

	admin := new(types.Admin)
	err := r.db.SqlDB.QueryRow(query, telegramID).Scan(
		&admin.ID,
		&admin.TelegramID,
		&admin.FirstName,
		&admin.LastName,
		&admin.Username,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admin by id: %w", err)
	}

	return admin, nil
}

func (r *AdminRepository) Insert(admin *types.Admin) (*types.Admin, error) {
	if admin == nil {
		return nil, fmt.Errorf("admin is nil")
	}

	query := `
	INSERT INTO admins 
	(telegram_id, first_name, last_name, username, created_at, updated_at)
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	result, err := r.db.SqlDB.Exec(
		query,
		admin.TelegramID,
		admin.FirstName,
		admin.LastName,
		admin.Username,
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
