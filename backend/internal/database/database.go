package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// DB wraps the SQL database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection
// Supports both MySQL DSN (mysql://...) and legacy SQLite path for backwards compatibility
func New(dsn string) (*DB, error) {
	var db *sql.DB
	var err error

	// Detect database type from DSN
	if strings.HasPrefix(dsn, "mysql://") {
		// MySQL DSN format: mysql://user:pass@host:port/dbname?parseTime=true
		// Convert to Go MySQL driver format: user:pass@tcp(host:port)/dbname?parseTime=true
		dsn = strings.TrimPrefix(dsn, "mysql://")

		// Parse the DSN to add tcp() wrapper around host:port
		// Format: user:pass@host:port/dbname -> user:pass@tcp(host:port)/dbname
		parts := strings.SplitN(dsn, "@", 2)
		if len(parts) == 2 {
			hostAndRest := parts[1]
			// Find the '/' that separates host:port from dbname
			slashIdx := strings.Index(hostAndRest, "/")
			if slashIdx > 0 {
				host := hostAndRest[:slashIdx]
				rest := hostAndRest[slashIdx:]
				dsn = parts[0] + "@tcp(" + host + ")" + rest
			}
		}

		db, err = sql.Open("mysql", dsn)
	} else {
		// Legacy SQLite path (for backwards compatibility during migration)
		return nil, fmt.Errorf("SQLite no longer supported - please use DATABASE_URL with MySQL DSN")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ MySQL database connected")

	return &DB{db}, nil
}

// Initialize creates all required tables
// NOTE: MySQL schema is created via migrations/001_initial_schema.sql on first run
// This function only runs additional migrations for schema evolution
func (db *DB) Initialize() error {
	log.Println("🔍 Checking database schema...")

	// Run migrations for existing databases
	if err := db.runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✅ Database initialized successfully")
	return nil
}

// runMigrations runs database migrations for schema updates
// Uses INFORMATION_SCHEMA to check for column existence (MySQL-compatible)
func (db *DB) runMigrations() error {
	dbName := os.Getenv("MYSQL_DATABASE")
	if dbName == "" {
		dbName = "claraverse" // default
	}

	// Helper function to check if column exists
	columnExists := func(tableName, columnName string) (bool, error) {
		var count int
		query := `
			SELECT COUNT(*)
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND COLUMN_NAME = ?
		`
		err := db.QueryRow(query, dbName, tableName, columnName).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}

	// Helper function to check if table exists
	tableExists := func(tableName string) (bool, error) {
		var count int
		query := `
			SELECT COUNT(*)
			FROM INFORMATION_SCHEMA.TABLES
			WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		`
		err := db.QueryRow(query, dbName, tableName).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}

	// Migration: Add audio_only column to providers table (if missing)
	if exists, _ := tableExists("providers"); exists {
		if colExists, _ := columnExists("providers", "audio_only"); !colExists {
			log.Println("📦 Running migration: Adding audio_only to providers table")
			if _, err := db.Exec("ALTER TABLE providers ADD COLUMN audio_only BOOLEAN DEFAULT FALSE"); err != nil {
				return fmt.Errorf("failed to add audio_only to providers: %w", err)
			}
			log.Println("✅ Migration completed: providers.audio_only added")
		}
	}

	// Migration: Add image_only column to providers table (if missing)
	if exists, _ := tableExists("providers"); exists {
		if colExists, _ := columnExists("providers", "image_only"); !colExists {
			log.Println("📦 Running migration: Adding image_only to providers table")
			if _, err := db.Exec("ALTER TABLE providers ADD COLUMN image_only BOOLEAN DEFAULT FALSE"); err != nil {
				return fmt.Errorf("failed to add image_only to providers: %w", err)
			}
			log.Println("✅ Migration completed: providers.image_only added")
		}
	}

	// Migration: Add image_edit_only column to providers table (if missing)
	if exists, _ := tableExists("providers"); exists {
		if colExists, _ := columnExists("providers", "image_edit_only"); !colExists {
			log.Println("📦 Running migration: Adding image_edit_only to providers table")
			if _, err := db.Exec("ALTER TABLE providers ADD COLUMN image_edit_only BOOLEAN DEFAULT FALSE"); err != nil {
				return fmt.Errorf("failed to add image_edit_only to providers: %w", err)
			}
			log.Println("✅ Migration completed: providers.image_edit_only added")
		}
	}

	// Migration: Add secure column to providers table (if missing)
	if exists, _ := tableExists("providers"); exists {
		if colExists, _ := columnExists("providers", "secure"); !colExists {
			log.Println("📦 Running migration: Adding secure to providers table")
			if _, err := db.Exec("ALTER TABLE providers ADD COLUMN secure BOOLEAN DEFAULT FALSE COMMENT 'Privacy-focused provider'"); err != nil {
				return fmt.Errorf("failed to add secure to providers: %w", err)
			}
			log.Println("✅ Migration completed: providers.secure added")
		}
	}

	// Migration: Add default_model column to providers table (if missing)
	if exists, _ := tableExists("providers"); exists {
		if colExists, _ := columnExists("providers", "default_model"); !colExists {
			log.Println("📦 Running migration: Adding default_model to providers table")
			if _, err := db.Exec("ALTER TABLE providers ADD COLUMN default_model VARCHAR(255)"); err != nil {
				return fmt.Errorf("failed to add default_model to providers: %w", err)
			}
			log.Println("✅ Migration completed: providers.default_model added")
		}
	}

	// Migration: Add smart_tool_router column to models table (if missing)
	if exists, _ := tableExists("models"); exists {
		if colExists, _ := columnExists("models", "smart_tool_router"); !colExists {
			log.Println("📦 Running migration: Adding smart_tool_router to models table")
			if _, err := db.Exec("ALTER TABLE models ADD COLUMN smart_tool_router BOOLEAN DEFAULT FALSE COMMENT 'Can predict tool usage'"); err != nil {
				return fmt.Errorf("failed to add smart_tool_router to models: %w", err)
			}
			log.Println("✅ Migration completed: models.smart_tool_router added")
		}
	}

	// Migration: Add agents_enabled column to models table (if missing)
	if exists, _ := tableExists("models"); exists {
		if colExists, _ := columnExists("models", "agents_enabled"); !colExists {
			log.Println("📦 Running migration: Adding agents_enabled to models table")
			if _, err := db.Exec("ALTER TABLE models ADD COLUMN agents_enabled BOOLEAN DEFAULT FALSE COMMENT 'Available in agent builder'"); err != nil {
				return fmt.Errorf("failed to add agents_enabled to models: %w", err)
			}
			log.Println("✅ Migration completed: models.agents_enabled added")
		}
	}

	// Migration: Add created_at and updated_at timestamps to models (if missing)
	if exists, _ := tableExists("models"); exists {
		if colExists, _ := columnExists("models", "created_at"); !colExists {
			log.Println("📦 Running migration: Adding created_at to models table")
			if _, err := db.Exec("ALTER TABLE models ADD COLUMN created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP"); err != nil {
				return fmt.Errorf("failed to add created_at to models: %w", err)
			}
			log.Println("✅ Migration completed: models.created_at added")
		}

		if colExists, _ := columnExists("models", "updated_at"); !colExists {
			log.Println("📦 Running migration: Adding updated_at to models table")
			if _, err := db.Exec("ALTER TABLE models ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"); err != nil {
				return fmt.Errorf("failed to add updated_at to models: %w", err)
			}
			log.Println("✅ Migration completed: models.updated_at added")
		}
	}

	// Migration: Add smart_tool_router column to model_aliases table (if missing)
	if exists, _ := tableExists("model_aliases"); exists {
		if colExists, _ := columnExists("model_aliases", "smart_tool_router"); !colExists {
			log.Println("📦 Running migration: Adding smart_tool_router to model_aliases table")
			if _, err := db.Exec("ALTER TABLE model_aliases ADD COLUMN smart_tool_router BOOLEAN DEFAULT FALSE"); err != nil {
				return fmt.Errorf("failed to add smart_tool_router to model_aliases: %w", err)
			}
			log.Println("✅ Migration completed: model_aliases.smart_tool_router added")
		}
	}

	log.Println("✅ All migrations completed")
	return nil
}
