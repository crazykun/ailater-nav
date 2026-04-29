package middleware

import (
	"ai-later-nav/internal/config"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"
)

type AccessLogEntry struct {
	Timestamp  string `json:"timestamp"`
	ClientIP   string `json:"client_ip"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	Latency    string `json:"latency"`
	UserAgent  string `json:"user_agent"`
	Referer    string `json:"referer"`
	RequestID  string `json:"request_id,omitempty"`
	Username   string `json:"username,omitempty"`
}

var accessLogger *json.Encoder

func InitAccessLogger() error {
	cfg := config.AppConfig.Log.AccessLog
	if !cfg.Enabled {
		return nil
	}

	logDir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	maxSize := cfg.MaxSize
	if maxSize <= 0 {
		maxSize = 100
	}
	maxBackups := cfg.MaxBackups
	if maxBackups <= 0 {
		maxBackups = 7
	}
	maxAge := cfg.MaxAge
	if maxAge <= 0 {
		maxAge = 7
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   false,
	}

	accessLogger = json.NewEncoder(lumberjackLogger)
	return nil
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.AppConfig.Log.AccessLog.Enabled || accessLogger == nil {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/admin") || strings.HasPrefix(path, "/static") {
			c.Next()
			return
		}

		start := time.Now()

		c.Next()

		entry := AccessLogEntry{
			Timestamp:  time.Now().Format(time.RFC3339),
			ClientIP:   c.ClientIP(),
			Method:     c.Request.Method,
			Path:       path,
			StatusCode: c.Writer.Status(),
			Latency:    time.Since(start).String(),
			UserAgent:  c.Request.UserAgent(),
			Referer:    c.Request.Referer(),
			RequestID:  c.GetHeader("X-Request-ID"),
		}

		if username, exists := c.Get("username"); exists {
			if u, ok := username.(string); ok {
				entry.Username = u
			}
		}

		accessLogger.Encode(entry)
	}
}
