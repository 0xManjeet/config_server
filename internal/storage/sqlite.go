package storage

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS data (
			key TEXT PRIMARY KEY,
			value JSON NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Get(key string) (json.RawMessage, error) {
	var value json.RawMessage
	err := s.db.QueryRow("SELECT value FROM data WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return value, err
}

func (s *Storage) Set(key string, value json.RawMessage) error {
	_, err := s.db.Exec(`
		INSERT INTO data (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
} 