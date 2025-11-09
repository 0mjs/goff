package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type FlagStore struct {
	db *sql.DB
}

type Flag struct {
	Key      string `json:"key"`
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`
	Variants string `json:"variants"` // JSON string
	Rules    string `json:"rules"`    // JSON string
	Default  string `json:"default"`  // JSON string
}

func NewFlagStore(dbPath string) (*FlagStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	store := &FlagStore{db: db}
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return store, nil
}

func (s *FlagStore) Close() error {
	return s.db.Close()
}

func (s *FlagStore) initSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS flags (
		key TEXT PRIMARY KEY,
		enabled INTEGER NOT NULL DEFAULT 1,
		type TEXT NOT NULL CHECK(type IN ('bool', 'string')),
		variants TEXT NOT NULL DEFAULT '{}',
		rules TEXT NOT NULL DEFAULT '[]',
		default_value TEXT NOT NULL
	);
	`
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	// Insert initial flags if table is empty
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM flags").Scan(&count); err != nil {
		return fmt.Errorf("check count: %w", err)
	}

	if count == 0 {
		if err := s.seedFlags(); err != nil {
			return fmt.Errorf("seed flags: %w", err)
		}
	}

	return nil
}

func (s *FlagStore) seedFlags() error {
	flags := []Flag{
		{
			Key:      "new_checkout",
			Enabled:  true,
			Type:     "bool",
			Variants: `{"true": 50, "false": 50}`,
			Rules: `[{
				"when": {"all": [{"attr": "plan", "op": "eq", "value": "pro"}]},
				"then": {"variants": {"true": 90, "false": 10}}
			}]`,
			Default: "false",
		},
		{
			Key:      "checkout_theme",
			Enabled:  true,
			Type:     "string",
			Variants: `{"red": 40, "blue": 30, "green": 30}`,
			Rules: `[{
				"when": {"all": [{"attr": "theme", "op": "eq", "value": "dark"}]},
				"then": {"variants": {"black": 100}}
			}]`,
			Default: `"red"`,
		},
		{
			Key:      "enable_logging",
			Enabled:  true,
			Type:     "bool",
			Variants: `{"true": 100, "false": 0}`,
			Rules:    `[]`,
			Default:  "true",
		},
		{
			Key:      "log_level",
			Enabled:  true,
			Type:     "string",
			Variants: `{"debug": 30, "info": 50, "warn": 15, "error": 5}`,
			Rules:    `[]`,
			Default:  `"info"`,
		},
	}

	for _, flag := range flags {
		if err := s.CreateFlag(flag); err != nil {
			return fmt.Errorf("seed flag %s: %w", flag.Key, err)
		}
	}

	log.Println("Seeded initial flags")
	return nil
}

func (s *FlagStore) GetAllFlags() ([]Flag, error) {
	rows, err := s.db.Query("SELECT key, enabled, type, variants, rules, default_value FROM flags ORDER BY key")
	if err != nil {
		return nil, fmt.Errorf("query flags: %w", err)
	}
	defer rows.Close()

	var flags []Flag
	for rows.Next() {
		var f Flag
		if err := rows.Scan(&f.Key, &f.Enabled, &f.Type, &f.Variants, &f.Rules, &f.Default); err != nil {
			return nil, fmt.Errorf("scan flag: %w", err)
		}
		flags = append(flags, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return flags, nil
}

func (s *FlagStore) GetFlag(key string) (*Flag, error) {
	var f Flag
	err := s.db.QueryRow(
		"SELECT key, enabled, type, variants, rules, default_value FROM flags WHERE key = ?",
		key,
	).Scan(&f.Key, &f.Enabled, &f.Type, &f.Variants, &f.Rules, &f.Default)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query flag: %w", err)
	}

	return &f, nil
}

func (s *FlagStore) CreateFlag(flag Flag) error {
	_, err := s.db.Exec(
		"INSERT INTO flags (key, enabled, type, variants, rules, default_value) VALUES (?, ?, ?, ?, ?, ?)",
		flag.Key, flag.Enabled, flag.Type, flag.Variants, flag.Rules, flag.Default,
	)
	if err != nil {
		return fmt.Errorf("insert flag: %w", err)
	}
	return nil
}

func (s *FlagStore) UpdateFlag(key string, flag Flag) error {
	result, err := s.db.Exec(
		"UPDATE flags SET enabled = ?, type = ?, variants = ?, rules = ?, default_value = ? WHERE key = ?",
		flag.Enabled, flag.Type, flag.Variants, flag.Rules, flag.Default, key,
	)
	if err != nil {
		return fmt.Errorf("update flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("flag not found: %s", key)
	}

	return nil
}

func (s *FlagStore) DeleteFlag(key string) error {
	result, err := s.db.Exec("DELETE FROM flags WHERE key = ?", key)
	if err != nil {
		return fmt.Errorf("delete flag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("flag not found: %s", key)
	}

	return nil
}

// ValidateJSON validates that a string is valid JSON
func ValidateJSON(s string) error {
	var v interface{}
	return json.Unmarshal([]byte(s), &v)
}
