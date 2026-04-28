package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port  string      `yaml:"port"`
	MySQL MySQLConfig `yaml:"mysql"`
	JWT   JWTConfig   `yaml:"jwt"`
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
		return overrideFromEnv()
	}

	if err := loadConfigFile("config.demo.yaml"); err == nil {
		return overrideFromEnv()
	}

	AppConfig = Config{
		Port: "8080",
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
	return overrideFromEnv()
}

func LoadRequiredConfig(filePath string) error {
	if err := loadConfigFile(filePath); err != nil {
		return err
	}
	return overrideFromEnv()
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

func overrideFromEnv() error {
	if port := os.Getenv("PORT"); port != "" {
		AppConfig.Port = port
	}

	if host := os.Getenv("MYSQL_HOST"); host != "" {
		AppConfig.MySQL.Host = host
	}

	if port := os.Getenv("MYSQL_PORT"); port != "" {
		parsed, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("invalid MYSQL_PORT %q: %w", port, err)
		}
		AppConfig.MySQL.Port = parsed
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

	return nil
}
