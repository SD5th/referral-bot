package storage

import (
	"database/sql"
	"fmt"
	"referral-bot/internal/core"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	core  *core.Core
	SqlDB *sql.DB
}

func NewDatabase(core *core.Core, dsn string) (*Database, error) {
	if core == nil {
		return nil, fmt.Errorf("core cannot be nil")
	}
	database := &Database{core: core}

	log := core.GetLogger()

	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	database.SqlDB = sqlDB

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	_, err = database.SqlDB.Exec(`
		PRAGMA foreign_keys = ON;
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to set pragmas: %w", err)
	}

	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Info("Database connected")
	return database, nil
}

func (db *Database) createTables() error {
	log := db.core.GetLogger()

	activeChannelTable :=
		`
		CREATE TABLE IF NOT EXISTS active_channel (
			id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),

			telegram_id BIGINT NOT NULL UNIQUE,

			type VARCHAR(10) NOT NULL CHECK (type IN ('private', 'group', 'supergroup', 'channel')),
			
			username VARCHAR(255),
			title VARCHAR(255),
			
			invite_link VARCHAR(100),
			
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
		`
	usersTable :=
		`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			
			telegram_id BIGINT NOT NULL UNIQUE,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255),
			username VARCHAR(255),
			member_status VARCHAR(13) NOT NULL CHECK (member_status IN ('creator', 'administrator', 'member', 'restricted', 'left', 'kicked')),
			
			invited_by_user_id BIGINT,
			invited_by_link_id BIGINT,
			
			invite_link_id BIGINT,
			
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			
			FOREIGN KEY (invited_by_user_id) REFERENCES users(id),
			FOREIGN KEY (invited_by_link_id) REFERENCES invite_links(id)
			)
		`
	adminsTable :=
		`
		CREATE TABLE IF NOT EXISTS admins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			
			telegram_id BIGINT NOT NULL UNIQUE,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255),
			username VARCHAR(255),
			
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`
	inviteLinksTable :=
		`
			CREATE TABLE IF NOT EXISTS invite_links (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				
				requester_id INTEGER NOT NULL,
				
				url VARCHAR(100) NOT NULL UNIQUE,
				name VARCHAR(255) NOT NULL UNIQUE,
				
				unique_joins INTEGER DEFAULT 0,
				
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				
				FOREIGN KEY (requester_id) REFERENCES users(id)
			)
		`
	channelActivityTable :=
		`
			CREATE TABLE IF NOT EXISTS channel_activity (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				channel_telegram_id INTEGER NOT NULL,
				
				user_telegram_id INTEGER NOT NULL,
				user_first_name VARCHAR(255) NOT NULL,
				user_last_name VARCHAR(255),
				user_username VARCHAR(255),

				inviter_telegram_id INTEGER,
				inviter_first_name VARCHAR(255),
				inviter_last_name VARCHAR(255),
				inviter_username VARCHAR(255),

				invite_link_url VARCHAR(100),
				invite_link_name VARCHAR(255) NOT NULL,
				
				old_status VARCHAR(13) NOT NULL CHECK (old_status IN ('creator', 'administrator', 'member', 'restricted', 'left', 'kicked')),
				new_status VARCHAR(13) NOT NULL CHECK (new_status IN ('creator', 'administrator', 'member', 'restricted', 'left', 'kicked')),
				
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`

	if _, err := db.SqlDB.Exec(activeChannelTable); err != nil {
		return fmt.Errorf("failed to create activeChannelTable: %w", err)
	}
	if _, err := db.SqlDB.Exec(usersTable); err != nil {
		return fmt.Errorf("failed to create usersTable: %w", err)
	}
	if _, err := db.SqlDB.Exec(adminsTable); err != nil {
		return fmt.Errorf("failed to create adminsTable: %w", err)
	}
	if _, err := db.SqlDB.Exec(inviteLinksTable); err != nil {
		return fmt.Errorf("failed to create inviteLinksTable: %w", err)
	}
	if _, err := db.SqlDB.Exec(channelActivityTable); err != nil {
		return fmt.Errorf("failed to create channelActivityTable: %w", err)
	}

	usersIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id)",
	}
	inviteLinksIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_invite_links_name ON invite_links(name)",
	}

	for _, indexSQL := range usersIndexes {
		if _, err := db.SqlDB.Exec(indexSQL); err != nil {
			log.Warn("Failed to create usersIndexes (non-critical): %v", err)
		}
	}

	for _, indexSQL := range inviteLinksIndexes {
		if _, err := db.SqlDB.Exec(indexSQL); err != nil {
			log.Warn("Failed to create inviteLinksIndexes (non-critical): %v", err)
		}
	}

	log.Info("All tables and indexes are created")
	return nil
}

func (d *Database) Close() error {
	if d.SqlDB != nil {
		return d.SqlDB.Close()
	}
	return nil
}
