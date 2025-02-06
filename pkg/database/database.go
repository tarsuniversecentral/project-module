package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tarsuniversecentral/project-module/config"
	"github.com/tarsuniversecentral/project-module/pkg/database/migration"
)

// InitDatabase initializes the database connection, configures the connection pool,
// verifies the connection, and runs migrations.
func InitDatabase() (*sql.DB, error) {
	// Load the configuration.
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Build the MySQL connection string.
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// Open the database connection.
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify the database connection with a ping.
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database")

	// Configure the database connection pool.
	db.SetMaxIdleConns(10)                 // Maximum number of idle connections.
	db.SetMaxOpenConns(100)                // Maximum number of open connections.
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum time a connection can be reused.

	// Run database migrations.
	if err = migration.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations applied successfully")
	return db, nil
}
