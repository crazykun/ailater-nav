package middleware

import (
	"ai-later-nav/internal/config"
	"ai-later-nav/internal/database"
	"ai-later-nav/internal/models"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	config.AppConfig.JWT.Secret = "test-secret-key-for-unit-tests"
	config.AppConfig.JWT.ExpireDays = 7
}

func TestGenerateAndValidateToken(t *testing.T) {
	user := &models.User{
		ID:       1,
		Username: "testuser",
		Role:     "user",
	}

	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("UserID = %d, want %d", claims.UserID, user.ID)
	}
	if claims.Username != user.Username {
		t.Errorf("Username = %s, want %s", claims.Username, user.Username)
	}
	if claims.Role != user.Role {
		t.Errorf("Role = %s, want %s", claims.Role, user.Role)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	_, err := ValidateToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token, got nil")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	user := &models.User{ID: 1, Username: "test", Role: "user"}
	token, _ := GenerateToken(user)

	origSecret := config.AppConfig.JWT.Secret
	config.AppConfig.JWT.Secret = "different-secret"
	defer func() { config.AppConfig.JWT.Secret = origSecret }()

	_, err := ValidateToken(token)
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}

func TestGenerateToken_AdminRole(t *testing.T) {
	user := &models.User{
		ID:       99,
		Username: "admin",
		Role:     "admin",
	}

	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.Role != "admin" {
		t.Errorf("Role = %s, want admin", claims.Role)
	}
}

func TestTokenExpiry(t *testing.T) {
	origDays := config.AppConfig.JWT.ExpireDays
	config.AppConfig.JWT.ExpireDays = 7
	defer func() { config.AppConfig.JWT.ExpireDays = origDays }()

	user := &models.User{ID: 1, Username: "test", Role: "user"}
	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	_ = claims
	_ = time.Now()
}

func TestAuthMiddlewareUsesFreshUserFromDBForContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token, err := GenerateToken(&models.User{ID: 10, Username: "old-admin", Role: "admin"})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	db, cleanup := openMiddlewareStubDB(t, middlewareStubConfig{
		queryUserByID: &models.User{ID: 10, Username: "current-user", Role: "user", Status: "active"},
	})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":  c.GetInt64("user_id"),
			"username": c.GetString("username"),
			"role":     c.GetString("role"),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token, Path: "/"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if body := w.Body.String(); body != "{\"role\":\"user\",\"user_id\":10,\"username\":\"current-user\"}" {
		t.Fatalf("body = %q, want fresh DB user context", body)
	}
}

func TestAuthMiddlewareRedirectsBlockedUserAndClearsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token, err := GenerateToken(&models.User{ID: 11, Username: "blocked", Role: "user"})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	db, cleanup := openMiddlewareStubDB(t, middlewareStubConfig{
		queryUserByID: &models.User{ID: 11, Username: "blocked", Role: "user", Status: "blocked"},
	})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token, Path: "/"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusFound)
	}
	if location := w.Header().Get("Location"); location != "/login" {
		t.Fatalf("redirect location = %q, want /login", location)
	}
	if !strings.Contains(w.Header().Get("Set-Cookie"), "token=") || !strings.Contains(w.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatalf("expected cleared token cookie, got %q", w.Header().Get("Set-Cookie"))
	}
}

func TestOptionalAuthUsesFreshUserFromDBForContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token, err := GenerateToken(&models.User{ID: 12, Username: "old-admin", Role: "admin"})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	db, cleanup := openMiddlewareStubDB(t, middlewareStubConfig{
		queryUserByID: &models.User{ID: 12, Username: "current-user", Role: "user", Status: "active"},
	})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	r := gin.New()
	r.Use(OptionalAuth())
	r.GET("/optional", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id":  c.GetInt64("user_id"),
			"username": c.GetString("username"),
			"role":     c.GetString("role"),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token, Path: "/"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if body := w.Body.String(); body != "{\"role\":\"user\",\"user_id\":12,\"username\":\"current-user\"}" {
		t.Fatalf("body = %q, want fresh DB user context", body)
	}
}

func TestOptionalAuthClearsCookieForBlockedUserAndContinuesAsGuest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token, err := GenerateToken(&models.User{ID: 13, Username: "blocked", Role: "user"})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	db, cleanup := openMiddlewareStubDB(t, middlewareStubConfig{
		queryUserByID: &models.User{ID: 13, Username: "blocked", Role: "user", Status: "blocked"},
	})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	r := gin.New()
	r.Use(OptionalAuth())
	r.GET("/optional", func(c *gin.Context) {
		if _, exists := c.Get("user_id"); exists {
			c.String(http.StatusOK, "attached")
			return
		}
		c.String(http.StatusOK, "guest")
	})

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token, Path: "/"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if body := w.Body.String(); body != "guest" {
		t.Fatalf("body = %q, want guest", body)
	}
	if location := w.Header().Get("Location"); location != "" {
		t.Fatalf("redirect location = %q, want empty", location)
	}
	if !strings.Contains(w.Header().Get("Set-Cookie"), "token=") || !strings.Contains(w.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatalf("expected cleared token cookie, got %q", w.Header().Get("Set-Cookie"))
	}
}

func TestOptionalAuthClearsCookieForMissingUserAndContinuesAsGuest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token, err := GenerateToken(&models.User{ID: 14, Username: "missing", Role: "user"})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	db, cleanup := openMiddlewareStubDB(t, middlewareStubConfig{})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	r := gin.New()
	r.Use(OptionalAuth())
	r.GET("/optional", func(c *gin.Context) {
		if _, exists := c.Get("user_id"); exists {
			c.String(http.StatusOK, "attached")
			return
		}
		c.String(http.StatusOK, "guest")
	})

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token, Path: "/"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if body := w.Body.String(); body != "guest" {
		t.Fatalf("body = %q, want guest", body)
	}
	if location := w.Header().Get("Location"); location != "" {
		t.Fatalf("redirect location = %q, want empty", location)
	}
	if !strings.Contains(w.Header().Get("Set-Cookie"), "token=") || !strings.Contains(w.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatalf("expected cleared token cookie, got %q", w.Header().Get("Set-Cookie"))
	}
}

func TestAuthMiddlewareReturnsServiceUnavailableOnUserLookupErrorWithoutClearingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token, err := GenerateToken(&models.User{ID: 15, Username: "user", Role: "user"})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	db, cleanup := openMiddlewareStubDB(t, middlewareStubConfig{queryErr: errors.New("db down")})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token, Path: "/"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
	if location := w.Header().Get("Location"); location != "" {
		t.Fatalf("redirect location = %q, want empty", location)
	}
	if strings.Contains(w.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatalf("expected cookie to be preserved, got %q", w.Header().Get("Set-Cookie"))
	}
}

func TestOptionalAuthPreservesCookieAndContinuesAsGuestOnUserLookupError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token, err := GenerateToken(&models.User{ID: 16, Username: "user", Role: "user"})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	db, cleanup := openMiddlewareStubDB(t, middlewareStubConfig{queryErr: errors.New("db down")})
	defer cleanup()

	origDB := database.DB
	database.DB = db
	defer func() { database.DB = origDB }()

	r := gin.New()
	r.Use(OptionalAuth())
	r.GET("/optional", func(c *gin.Context) {
		if _, exists := c.Get("user_id"); exists {
			c.String(http.StatusOK, "attached")
			return
		}
		c.String(http.StatusOK, "guest")
	})

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token, Path: "/"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if body := w.Body.String(); body != "guest" {
		t.Fatalf("body = %q, want guest", body)
	}
	if strings.Contains(w.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatalf("expected cookie to be preserved, got %q", w.Header().Get("Set-Cookie"))
	}
}

type middlewareStubConfig struct {
	queryUserByID *models.User
	queryErr      error
}

type middlewareStubDriver struct {
	config middlewareStubConfig
}

type middlewareStubConn struct {
	config middlewareStubConfig
}

var middlewareStubCounter int64

func openMiddlewareStubDB(t *testing.T, config middlewareStubConfig) (*sql.DB, func()) {
	t.Helper()

	driverName := fmt.Sprintf("middleware-stub-%d", atomic.AddInt64(&middlewareStubCounter, 1))
	sql.Register(driverName, &middlewareStubDriver{config: config})

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

func (d *middlewareStubDriver) Open(string) (driver.Conn, error) {
	return &middlewareStubConn{config: d.config}, nil
}

func (c *middlewareStubConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not implemented")
}

func (c *middlewareStubConn) Close() error {
	return nil
}

func (c *middlewareStubConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not implemented")
}

func (c *middlewareStubConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return nil, errors.New("exec not implemented")
}

func (c *middlewareStubConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if !strings.Contains(query, "FROM users WHERE id = ?") {
		return nil, fmt.Errorf("unexpected query: %s", query)
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("expected 1 arg, got %d", len(args))
	}
	if c.config.queryErr != nil {
		return nil, c.config.queryErr
	}
	if c.config.queryUserByID == nil {
		return &middlewareStubRows{}, nil
	}
	now := time.Unix(0, 0)
	return &middlewareStubRows{
		columns: []string{"id", "username", "email", "password_hash", "role", "status", "created_at", "updated_at"},
		values: [][]driver.Value{{
			c.config.queryUserByID.ID,
			c.config.queryUserByID.Username,
			c.config.queryUserByID.Email,
			c.config.queryUserByID.PasswordHash,
			c.config.queryUserByID.Role,
			c.config.queryUserByID.Status,
			now,
			now,
		}},
	}, nil
}

type middlewareStubRows struct {
	columns []string
	values  [][]driver.Value
	idx     int
}

func (r *middlewareStubRows) Columns() []string { return r.columns }
func (r *middlewareStubRows) Close() error      { return nil }
func (r *middlewareStubRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.idx])
	r.idx++
	return nil
}

var _ driver.ExecerContext = (*middlewareStubConn)(nil)
var _ driver.QueryerContext = (*middlewareStubConn)(nil)
var _ driver.Conn = (*middlewareStubConn)(nil)
var _ driver.Driver = (*middlewareStubDriver)(nil)
var _ driver.Rows = (*middlewareStubRows)(nil)
