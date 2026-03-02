package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/masroof-maindak/pubsub/internal/db/queries"

	_ "github.com/mattn/go-sqlite3"
)

const dbname string = "broker.db"

var db *sql.DB

func InitTestDb() error {
	var err error

	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		return fmt.Errorf("failed to open test database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to verify test database connection: %w", err)
	}

	pragmas := []string{
		"PRAGMA foreign_keys=ON",
		"PRAGMA temp_store=MEMORY",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to set pragma '%s': %w", pragma, err)
		}
	}

	_, err = db.Exec(queries.SchemaCreationStatement)
	if err != nil {
		return fmt.Errorf("failed to create test schema: %w", err)
	}

	return nil
}

func CleanupTestDb() error {
	if db != nil {
		if err := db.Close(); err != nil {
			return fmt.Errorf("failed to close test database: %w", err)
		}
		db = nil
	}
	return nil
}

func InitDb(dir string) error {
	path := filepath.Join(dir, dbname)
	var err error

	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to verify database connection: %w", err)
	}

	pragmas := []string{
		"PRAGMA foreign_keys=ON",
		"PRAGMA temp_store=MEMORY",
		"PRAGMA mmap_size=4000000000", // 4 GB
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return fmt.Errorf("failed to set pragma '%s': %w", pragma, err)
		}
	}

	_, err = db.Exec(queries.SchemaCreationStatement)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func CleanupDb() error {
	if db != nil {
		if err := db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	return nil
}

func SaveLatestMessage(topic string, msg string) error {
	_, err := db.Exec(queries.UpdateLatestMsgMsgStatement, topic, msg)
	if err != nil {
		return fmt.Errorf("failed to save latest message: %w", err)
	}
	return nil
}

func GetLatestMessage(topic string) (string, error) {
	var msg sql.NullString
	err := db.QueryRow(queries.GetLatestMsgStatement, topic).Scan(&msg)
	if err == sql.ErrNoRows {
		return "", nil // No history yet
	}

	if err != nil {
		return "", fmt.Errorf("failed to get latest message: %w", err)
	}
	return msg.String, nil
}
