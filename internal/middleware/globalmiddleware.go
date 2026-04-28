package middleware

import (
	"ai-later-nav/internal/config"
	"ai-later-nav/internal/models"
	"net/http"
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

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func OptionalAuth() gin.HandlerFunc {
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

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			c.HTML(http.StatusForbidden, "error.html", gin.H{
				"error": "需要管理员权限",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func AddGlobalContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		copyright := config.AppConfig.Copyright
		if copyright == "" {
			copyright = "AI导航 © 2024"
		}
		c.Set("Copyright", copyright)

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
