package repository

import (
	"ai-later-nav/internal/database"
	"database/sql"
	"fmt"
)

type SettingRepository struct {
	db *sql.DB
}

func NewSettingRepository() *SettingRepository {
	return &SettingRepository{db: database.DB}
}

func (r *SettingRepository) Get(key string) (string, error) {
	var value string
	err := r.db.QueryRow("SELECT `value` FROM system_settings WHERE `key` = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get setting %s: %w", key, err)
	}
	return value, nil
}

func (r *SettingRepository) GetAll() (map[string]string, error) {
	rows, err := r.db.Query("SELECT `key`, `value` FROM system_settings")
	if err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		settings[k] = v
	}
	return settings, rows.Err()
}

func (r *SettingRepository) Set(key, value string) error {
	_, err := r.db.Exec(
		"INSERT INTO system_settings (`key`, `value`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `value` = VALUES(`value`)",
		key, value,
	)
	return err
}

func (r *SettingRepository) SetMultiple(settings map[string]string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO system_settings (`key`, `value`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `value` = VALUES(`value`)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for k, v := range settings {
		if _, err := stmt.Exec(k, v); err != nil {
			return fmt.Errorf("set setting %s: %w", k, err)
		}
	}
	return tx.Commit()
}

func (r *SettingRepository) SeedDefaults(defaults map[string]string) error {
	for k, v := range defaults {
		_, err := r.db.Exec(
			"INSERT IGNORE INTO system_settings (`key`, `value`) VALUES (?, ?)",
			k, v,
		)
		if err != nil {
			return fmt.Errorf("seed setting %s: %w", k, err)
		}
	}
	return nil
}
