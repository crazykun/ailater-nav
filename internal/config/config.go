package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port      string        `yaml:"port"`
	Copyright string        `yaml:"copyright"`
	Admin     AdminConfig   `yaml:"admin"`
	Session   SessionConfig `yaml:"session"`
	MySQL     MySQLConfig   `yaml:"mysql"`
	JWT       JWTConfig     `yaml:"jwt"`
}

type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type SessionConfig struct {
	Secret string `yaml:"secret"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireDays int    `yaml:"expire_days"`
}

var AppConfig Config

func LoadConfig() error {
	if err := loadConfigFile("config.yaml"); err == nil {
		overrideFromEnv()
		return nil
	}

	if err := loadConfigFile("config.demo.yaml"); err == nil {
		overrideFromEnv()
		return nil
	}

	AppConfig = Config{
		Port:      "8080",
		Copyright: "AI导航 © 2024",
		Admin: AdminConfig{
			Username: "admin",
			Password: "admin123",
		},
		Session: SessionConfig{
			Secret: "your-secret-key-here",
		},
		MySQL: MySQLConfig{
			Host:     "localhost",
			Port:     3306,
			Username: "root",
			Password: "",
			Database: "ai_later",
		},
		JWT: JWTConfig{
			Secret:     "your-jwt-secret-here",
			ExpireDays: 7,
		},
	}
	overrideFromEnv()
	return nil
}

func loadConfigFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	return decoder.Decode(&AppConfig)
}

func overrideFromEnv() {
	if port := os.Getenv("PORT"); port != "" {
		AppConfig.Port = port
	}

	if copyright := os.Getenv("COPYRIGHT"); copyright != "" {
		AppConfig.Copyright = copyright
	}

	if username := os.Getenv("ADMIN_USERNAME"); username != "" {
		AppConfig.Admin.Username = username
	}

	if password := os.Getenv("ADMIN_PASSWORD"); password != "" {
		AppConfig.Admin.Password = password
	}

	if secret := os.Getenv("SESSION_SECRET"); secret != "" {
		AppConfig.Session.Secret = secret
	}

	if host := os.Getenv("MYSQL_HOST"); host != "" {
		AppConfig.MySQL.Host = host
	}

	if port := os.Getenv("MYSQL_PORT"); port != "" {
		// parse port
		AppConfig.MySQL.Port = 3306
	}

	if user := os.Getenv("MYSQL_USER"); user != "" {
		AppConfig.MySQL.Username = user
	}

	if pass := os.Getenv("MYSQL_PASSWORD"); pass != "" {
		AppConfig.MySQL.Password = pass
	}

	if db := os.Getenv("MYSQL_DATABASE"); db != "" {
		AppConfig.MySQL.Database = db
	}

	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		AppConfig.JWT.Secret = secret
	}
}
