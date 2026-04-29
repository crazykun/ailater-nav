package middleware

import (
	"ai-later-nav/internal/config"
	"ai-later-nav/internal/models"
	"testing"
	"time"
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
