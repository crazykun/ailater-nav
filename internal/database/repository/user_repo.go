package repository

import (
	"ai-later-nav/internal/database"
	"ai-later-nav/internal/models"
	"database/sql"
	"fmt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{db: database.DB}
}

func (r *UserRepository) Create(user *models.User) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO users (username, password_hash, role)
		VALUES (?, ?, ?)
	`, user.Username, user.PasswordHash, user.Role)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return result.LastInsertId()
}

func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(`
		SELECT id, username, COALESCE(email, ''), password_hash, role, created_at, updated_at
		FROM users WHERE id = ?
	`, id).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(`
		SELECT id, username, COALESCE(email, ''), password_hash, role, created_at, updated_at
		FROM users WHERE username = ?
	`, username).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by username: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) UpdatePassword(userID int64, passwordHash string) error {
	_, err := r.db.Exec(`
		UPDATE users SET password_hash = ?, updated_at = NOW() WHERE id = ?
	`, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateRole(userID int64, role string) error {
	_, err := r.db.Exec("UPDATE users SET role = ?, updated_at = NOW() WHERE id = ?", role, userID)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

func (r *UserRepository) AddFavorite(userID, siteID int64) error {
	_, err := r.db.Exec("INSERT IGNORE INTO favorites (user_id, site_id) VALUES (?, ?)", userID, siteID)
	return err
}

func (r *UserRepository) RemoveFavorite(userID, siteID int64) error {
	_, err := r.db.Exec("DELETE FROM favorites WHERE user_id = ? AND site_id = ?", userID, siteID)
	return err
}

func (r *UserRepository) IsFavorite(userID, siteID int64) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM favorites WHERE user_id = ? AND site_id = ?", userID, siteID).Scan(&count)
	return count > 0, err
}

func (r *UserRepository) GetFavorites(userID int64) ([]int64, error) {
	rows, err := r.db.Query("SELECT site_id FROM favorites WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, fmt.Errorf("query favorites: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan favorite: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (r *UserRepository) GetAllUsers() ([]*models.User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, COALESCE(email, ''), password_hash, role, created_at, updated_at
		FROM users ORDER BY id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query all users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

func (r *UserRepository) CountTodayUsers() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURDATE()").Scan(&count)
	return count, err
}

func (r *UserRepository) GetRecentUsers(limit int) ([]*models.User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, COALESCE(email, ''), password_hash, role, created_at, updated_at
		FROM users ORDER BY created_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}
