package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // register MySQL driver
	"github.com/todo-app/backend/config"
	"github.com/todo-app/backend/logger"
)

// DB is the application-wide connection pool.
// After ConnectDatabase returns successfully this is safe to use concurrently.
var DB *sql.DB

// ConnectDatabase opens and validates a MySQL connection, then runs
// all startup DDL (CREATE TABLE IF NOT EXISTS + legacy cleanup).
func ConnectDatabase(cfg config.Config, log logger.Logger) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.GetDBUser(), cfg.GetDBPassword(), cfg.GetDBHost(), cfg.GetDBPort(), cfg.GetDBName())

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("failed to open MySQL connection", logger.F("error", err))
	}

	if err = sqlDB.PingContext(context.Background()); err != nil {
		log.Fatal("failed to ping MySQL — check credentials and host", logger.F("error", err))
	}

	if err = runMigrations(sqlDB, log); err != nil {
		log.Fatal("schema migration failed", logger.F("error", err))
	}

	DB = sqlDB
	log.Info("MySQL database connected and migrated successfully",
		logger.F("host", cfg.GetDBHost()),
		logger.F("db", cfg.GetDBName()),
	)
}

// CloseDatabase cleanly closes the MySQL connection pool
func CloseDatabase() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// runMigrations executes legacy cleanup and CREATE TABLE IF NOT EXISTS DDL.
// Errors on index creation are suppressed — MySQL raises an error if the
// index already exists, which is expected on subsequent startups.
func runMigrations(db *sql.DB, log logger.Logger) error {
	ctx := context.Background()

	// ── Legacy cleanup (idempotent — ignore errors if objects don't exist) ──
	legacyCleanup := []string{
		"ALTER TABLE todos DROP FOREIGN KEY fk_todos_groups",
		"ALTER TABLE todos DROP COLUMN group_id",
		"DROP TABLE IF EXISTS todo_groups",
	}
	for _, stmt := range legacyCleanup {
		_, _ = db.ExecContext(ctx, stmt) // intentionally ignore error
	}

	// ── Schema DDL ────────────────────────────────────────────────────────
	ddl := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id         INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name       VARCHAR(100) NOT NULL,
			email      VARCHAR(191) NOT NULL,
			password   VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY idx_users_email (email)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS todos (
			id             INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			title          VARCHAR(255) NOT NULL,
			description    TEXT NOT NULL DEFAULT '',
			completed      BOOLEAN NOT NULL DEFAULT FALSE,
			due_date       TIMESTAMP NULL DEFAULT NULL,
			user_id        INT UNSIGNED NOT NULL,
			parent_todo_id INT UNSIGNED NULL DEFAULT NULL,
			created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			CONSTRAINT fk_todos_users  FOREIGN KEY (user_id)        REFERENCES users (id) ON DELETE CASCADE,
			CONSTRAINT fk_todos_parent FOREIGN KEY (parent_todo_id) REFERENCES todos (id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS group_shares (
			id                   INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			group_id             INT UNSIGNED NOT NULL,
			owner_id             INT UNSIGNED NOT NULL,
			shared_with_user_id  INT UNSIGNED NOT NULL,
			permission           VARCHAR(50) NOT NULL,
			created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY idx_group_share_unique (group_id, shared_with_user_id),
			CONSTRAINT fk_group_shares_group  FOREIGN KEY (group_id)            REFERENCES todos (id) ON DELETE CASCADE,
			CONSTRAINT fk_group_shares_owner  FOREIGN KEY (owner_id)            REFERENCES users (id) ON DELETE CASCADE,
			CONSTRAINT fk_group_shares_target FOREIGN KEY (shared_with_user_id) REFERENCES users (id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, stmt := range ddl {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("DDL failed: %w", err)
		}
	}

	// ── Indexes (suppressed if already exist) ─────────────────────────────
	indexes := []string{
		"CREATE INDEX idx_todos_user_id   ON todos (user_id)",
		"CREATE INDEX idx_todos_completed  ON todos (completed)",
		"CREATE INDEX idx_todos_parent_id  ON todos (parent_todo_id)",
	}
	for _, stmt := range indexes {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			log.Debug("index already exists or creation skipped", logger.F("stmt", stmt))
		}
	}

	return nil
}
