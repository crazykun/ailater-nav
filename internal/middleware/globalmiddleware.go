package middleware

import (
	"ai-later-nav/internal/config"
	"ai-later-nav/internal/database/repository"
	"ai-later-nav/internal/models"
	"ai-later-nav/internal/services"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Duration(config.AppConfig.JWT.ExpireDays) * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWT.Secret))
}

func ValidateToken(tokenString string) (*models.UserClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int64(claims["user_id"].(float64))
		username := claims["username"].(string)
		role := claims["role"].(string)
		return &models.UserClaims{
			UserID:   userID,
			Username: username,
			Role:     role,
		}, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func AuthMiddleware() gin.HandlerFunc {
	userService := services.NewUserService()

	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		claims, err := ValidateToken(token)
		if err != nil {
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		user, err := userService.GetByID(claims.UserID)
		if err != nil {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		if user == nil || userService.IsBlocked(user) {
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)
		c.Next()
	}
}

func OptionalAuth() gin.HandlerFunc {
	userService := services.NewUserService()

	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil {
			c.Next()
			return
		}

		claims, err := ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		user, err := userService.GetByID(claims.UserID)
		if err != nil {
			c.Next()
			return
		}
		if user == nil || userService.IsBlocked(user) {
			c.SetCookie("token", "", -1, "/", "", false, true)
			c.Next()
			return
		}

		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("role", user.Role)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"error": "需要登录",
			})
			c.Abort()
			return
		}
		roleStr, ok := role.(string)
		if !ok || roleStr != "admin" {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"error": "需要管理员权限，当前角色: " + roleStr,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func AddGlobalContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		ss := services.GetSettingService()
		copyright := ss.GetSetting("copyright")
		c.Set("Copyright", copyright)
		c.Set("SiteName", ss.GetSetting("site_name"))

		if userID, exists := c.Get("user_id"); exists {
			c.Set("isLoggedIn", true)
			c.Set("userID", userID)
			c.Set("username", c.GetString("username"))
		} else {
			c.Set("isLoggedIn", false)
		}

		c.Next()
	}
}

func SetupRequired() gin.HandlerFunc {
	var (
		userRepo  *repository.UserRepository
		repoOnce  sync.Once
		setupDone atomic.Bool
	)

	return func(c *gin.Context) {
		if setupDone.Load() {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/static/") || path == "/setup" || path == "/api/setup" {
			c.Next()
			return
		}

		repoOnce.Do(func() {
			userRepo = repository.NewUserRepository()
		})

		count, err := userRepo.CountUsers()
		if err != nil {
			c.Next()
			return
		}

		if count == 0 {
			c.Redirect(http.StatusFound, "/setup")
			c.Abort()
			return
		}

		setupDone.Store(true)
		c.Next()
	}
}
