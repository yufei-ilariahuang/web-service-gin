package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB(dsn string) error {
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// ===== CRITICAL CONNECTION POOL SETTINGS =====

	// SetMaxOpenConns: Maximum number of open connections to the database
	// Set to 25 for RDS t3.micro (free tier limitation)
	// Too high = connection limit exceeded on server
	// Too low = connection waiting queue, slower performance
	DB.SetMaxOpenConns(25)
	// SetMaxIdleConns: Number of connections to keep idle
	// Set to 5 for efficient resource usage
	// Idle connections are ready for immediate use
	DB.SetMaxIdleConns(5)
	// SetConnMaxLifetime: Maximum time a connection can be reused
	// Set to 5 minutes - prevents stale connections
	// RDS closes connections after 900 seconds, so 5 min is safe
	DB.SetConnMaxLifetime(5 * time.Minute)
	// SetConnMaxIdleTime: Maximum idle time before connection closes
	// Set to 2 minutes - frees resources if not used
	DB.SetConnMaxIdleTime(2 * time.Minute)

	// Test connection immediately
	if err := DB.Ping(); err != nil {
		return err
	}

	fmt.Println("Database connected!")
	return nil
}

func CloseDB() error {
	return DB.Close()
}
