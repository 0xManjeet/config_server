package storage

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db    *sql.DB
	cache map[string]json.RawMessage
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

	// Initialize cache and load existing data
	storage := &Storage{
		db:    db,
		cache: make(map[string]json.RawMessage),
	}

	// Load existing data into cache
	rows, err := db.Query("SELECT key, value FROM data")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var value json.RawMessage
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		storage.cache[key] = value
	}

	return storage, nil
}

func (s *Storage) Get(key string) (json.RawMessage, error) {
	// First check cache
	if value, exists := s.cache[key]; exists {
		return value, nil
	}

	// If not in cache, check database
	var value json.RawMessage
	err := s.db.QueryRow("SELECT value FROM data WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Add to cache
	s.cache[key] = value
	return value, nil
}

func (s *Storage) Set(key string, value json.RawMessage) error {
	// Update cache immediately
	s.cache[key] = value

	// Async database update
	go func() {
		_, err := s.db.Exec(`
			INSERT INTO data (key, value) VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value
		`, key, value)
		if err != nil {
			println("Error writing to database:", err.Error())
		}
	}()

	return nil
}
