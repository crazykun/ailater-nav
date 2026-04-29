package services

import (
	"ai-later-nav/internal/database"
	"ai-later-nav/internal/models"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestIsBlocked(t *testing.T) {
	tests := []struct {
		name     string
		user     *models.User
		expected bool
	}{
		{name: "blocked user returns true", user: &models.User{Status: "blocked"}, expected: true},
		{name: "active user returns false", user: &models.User{Status: "active"}, expected: false},
		{name: "empty status returns false", user: &models.User{Status: ""}, expected: false},
	}

	svc := NewUserService()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.IsBlocked(tt.user); got != tt.expected {
				t.Fatalf("IsBlocked() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBlockedUserCannotLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}

	db, cleanup := openAdminUserServiceStubDB(t, adminUserServiceStubConfig{
		queryUserByUsername: &models.User{
			ID:           7,
			Username:     "blocked-user",
			PasswordHash: string(hash),
			Role:         "user",
			Status:       "blocked",
		},
	})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	svc := NewUserService()
	_, err = svc.Login("blocked-user", "secret123")
	if !errors.Is(err, ErrUserBlocked) {
		t.Fatalf("Login() error = %v, want %v", err, ErrUserBlocked)
	}
}

func TestResetPasswordByAdminUpdatesPasswordWithoutOldPasswordCheck(t *testing.T) {
	db, cleanup := openAdminUserServiceStubDB(t, adminUserServiceStubConfig{
		queryUserByID: &models.User{ID: 9, Username: "target", Role: "user", Status: "active"},
	})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	svc := NewUserService()
	if err := svc.ResetPasswordByAdmin(9, "new-password-123"); err != nil {
		t.Fatalf("ResetPasswordByAdmin() error = %v", err)
	}

	if got := adminUserServiceLastPasswordHash.Load(); got == nil {
		t.Fatal("expected password update to be called")
	} else if err := bcrypt.CompareHashAndPassword([]byte(got.(string)), []byte("new-password-123")); err != nil {
		t.Fatalf("updated hash does not match new password: %v", err)
	}
}

func TestSetUserRoleRejectsSelfModification(t *testing.T) {
	svc := NewUserService()

	err := svc.SetUserRole(3, 3, "user")
	if !errors.Is(err, ErrCannotModifySelfRole) {
		t.Fatalf("SetUserRole() error = %v, want %v", err, ErrCannotModifySelfRole)
	}
}

func TestSetUserRoleUpdatesTargetRole(t *testing.T) {
	db, cleanup := openAdminUserServiceStubDB(t, adminUserServiceStubConfig{
		queryUserByID: &models.User{ID: 5, Username: "member", Role: "user", Status: "active"},
	})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	svc := NewUserService()
	if err := svc.SetUserRole(1, 5, "admin"); err != nil {
		t.Fatalf("SetUserRole() error = %v", err)
	}

	if got := adminUserServiceLastRole.Load(); got != "admin" {
		t.Fatalf("updated role = %v, want admin", got)
	}
}

func TestSetUserStatusRejectsBlockingSelf(t *testing.T) {
	svc := NewUserService()

	err := svc.SetUserStatus(8, 8, "blocked")
	if !errors.Is(err, ErrCannotBlockSelf) {
		t.Fatalf("SetUserStatus() error = %v, want %v", err, ErrCannotBlockSelf)
	}
}

func TestSetUserStatusUpdatesTargetStatus(t *testing.T) {
	db, cleanup := openAdminUserServiceStubDB(t, adminUserServiceStubConfig{
		queryUserByID: &models.User{ID: 6, Username: "member", Role: "user", Status: "active"},
	})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	svc := NewUserService()
	if err := svc.SetUserStatus(1, 6, "blocked"); err != nil {
		t.Fatalf("SetUserStatus() error = %v", err)
	}

	if got := adminUserServiceLastStatus.Load(); got != "blocked" {
		t.Fatalf("updated status = %v, want blocked", got)
	}
}

func TestSetUserRoleRejectsInvalidRole(t *testing.T) {
	db, cleanup := openAdminUserServiceStubDB(t, adminUserServiceStubConfig{
		queryUserByID: &models.User{ID: 5, Username: "member", Role: "user", Status: "active"},
	})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	svc := NewUserService()
	err := svc.SetUserRole(1, 5, "owner")
	if err == nil {
		t.Fatal("expected error for invalid role, got nil")
	}
	if !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected invalid role error, got %v", err)
	}
}

func TestSetUserStatusRejectsInvalidStatus(t *testing.T) {
	db, cleanup := openAdminUserServiceStubDB(t, adminUserServiceStubConfig{
		queryUserByID: &models.User{ID: 6, Username: "member", Role: "user", Status: "active"},
	})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	svc := NewUserService()
	err := svc.SetUserStatus(1, 6, "disabled")
	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

type adminUserServiceStubConfig struct {
	queryUserByID       *models.User
	queryUserByUsername *models.User
}

type adminUserServiceStubDriver struct {
	config adminUserServiceStubConfig
}

type adminUserServiceStubConn struct {
	config adminUserServiceStubConfig
}

var adminUserServiceStubCounter int64
var adminUserServiceLastPasswordHash atomic.Value
var adminUserServiceLastRole atomic.Value
var adminUserServiceLastStatus atomic.Value

func openAdminUserServiceStubDB(t *testing.T, config adminUserServiceStubConfig) (*sql.DB, func()) {
	t.Helper()

	adminUserServiceLastPasswordHash = atomic.Value{}
	adminUserServiceLastRole = atomic.Value{}
	adminUserServiceLastStatus = atomic.Value{}

	driverName := fmt.Sprintf("admin-user-service-stub-%d", atomic.AddInt64(&adminUserServiceStubCounter, 1))
	sql.Register(driverName, &adminUserServiceStubDriver{config: config})

	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatalf("open stub db: %v", err)
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close stub db: %v", err)
		}
	}

	return db, cleanup
}

func (d *adminUserServiceStubDriver) Open(string) (driver.Conn, error) {
	return &adminUserServiceStubConn{config: d.config}, nil
}

func (c *adminUserServiceStubConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not implemented")
}

func (c *adminUserServiceStubConn) Close() error {
	return nil
}

func (c *adminUserServiceStubConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not implemented")
}

func (c *adminUserServiceStubConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	normalizedQuery := strings.Join(strings.Fields(query), " ")

	switch normalizedQuery {
	case "UPDATE users SET password_hash = ?, updated_at = NOW() WHERE id = ?":
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 args for password update, got %d", len(args))
		}
		adminUserServiceLastPasswordHash.Store(args[0].Value.(string))
		return driver.RowsAffected(1), nil
	case "UPDATE users SET role = ?, updated_at = NOW() WHERE id = ?":
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 args for role update, got %d", len(args))
		}
		adminUserServiceLastRole.Store(args[0].Value.(string))
		return driver.RowsAffected(1), nil
	case "UPDATE users SET status = ?, updated_at = NOW() WHERE id = ?":
		if len(args) != 2 {
			return nil, fmt.Errorf("expected 2 args for status update, got %d", len(args))
		}
		adminUserServiceLastStatus.Store(args[0].Value.(string))
		return driver.RowsAffected(1), nil
	default:
		return nil, fmt.Errorf("unexpected exec query: %s", query)
	}
}

func (c *adminUserServiceStubConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "FROM users WHERE username = ?"):
		if c.config.queryUserByUsername == nil {
			return &adminUserServiceStubRows{}, nil
		}
		return singleUserRows(c.config.queryUserByUsername), nil
	case strings.Contains(query, "FROM users WHERE id = ?"):
		if c.config.queryUserByID == nil {
			return &adminUserServiceStubRows{}, nil
		}
		if len(args) != 1 {
			return nil, fmt.Errorf("expected 1 arg for user lookup, got %d", len(args))
		}
		return singleUserRows(c.config.queryUserByID), nil
	default:
		return nil, fmt.Errorf("unexpected query: %s", query)
	}
}

type adminUserServiceStubRows struct {
	columns []string
	values  [][]driver.Value
	idx     int
}

func singleUserRows(user *models.User) driver.Rows {
	now := time.Unix(0, 0)

	return &adminUserServiceStubRows{
		columns: []string{"id", "username", "email", "password_hash", "role", "status", "created_at", "updated_at"},
		values: [][]driver.Value{{
			user.ID,
			user.Username,
			user.Email,
			user.PasswordHash,
			user.Role,
			user.Status,
			now,
			now,
		}},
	}
}

func (r *adminUserServiceStubRows) Columns() []string { return r.columns }
func (r *adminUserServiceStubRows) Close() error      { return nil }
func (r *adminUserServiceStubRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.idx])
	r.idx++
	return nil
}

var _ driver.ExecerContext = (*adminUserServiceStubConn)(nil)
var _ driver.QueryerContext = (*adminUserServiceStubConn)(nil)
var _ driver.Conn = (*adminUserServiceStubConn)(nil)
var _ driver.Driver = (*adminUserServiceStubDriver)(nil)
var _ driver.Rows = (*adminUserServiceStubRows)(nil)
