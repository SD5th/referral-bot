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
		SELECT 
			id, 
			telegram_id, first_name, last_name, username, 
			created_at, updated_at
		FROM admins 
		WHERE 
			id = ?
	`

	var admin types.Admin
	err := r.db.SqlDB.QueryRow(query, id).Scan(
		&admin.ID,
		&admin.TelegramID, &admin.FirstName, &admin.LastName, &admin.Username,
		&admin.CreatedAt, &admin.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &admin, nil
}

func (r *AdminRepository) GetByTelegramID(telegramID int64) (*types.Admin, error) {
	query := `
		SELECT 
			id, 
			telegram_id, first_name, last_name, username, 
			created_at, updated_at
		FROM admins 
		WHERE 
			telegram_id = ?
	`

	var admin types.Admin
	err := r.db.SqlDB.QueryRow(query, telegramID).Scan(
		&admin.ID,
		&admin.TelegramID, &admin.FirstName, &admin.LastName, &admin.Username,
		&admin.CreatedAt, &admin.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &admin, nil
}

func (r *AdminRepository) Insert(admin *types.Admin) (*types.Admin, error) {
	if admin == nil {
		return nil, fmt.Errorf("admin is nil")
	}

	query := `
		INSERT INTO admins (
			telegram_id, first_name, last_name, username, 
			created_at, updated_at
		)
		VALUES (
			?, ?, ?, ?, 
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		RETURNING 
			id, 
			telegram_id, first_name, last_name, username, 
			created_at, updated_at
	`
	var insertedAdmin types.Admin
	err := r.db.SqlDB.QueryRow(query,
		admin.TelegramID, admin.FirstName, admin.LastName, admin.Username,
	).Scan(
		&insertedAdmin.ID,
		&insertedAdmin.TelegramID, &insertedAdmin.FirstName, &insertedAdmin.LastName, &insertedAdmin.Username,
		&insertedAdmin.CreatedAt, &insertedAdmin.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &insertedAdmin, nil
}
