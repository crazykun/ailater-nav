package main

import (
	"ai-later-nav/internal/config"
	"database/sql"
	"fmt"
	"log"

	mysql "github.com/go-sql-driver/mysql"
)

func main() {
	if err := config.LoadRequiredConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg := mysql.NewConfig()
	cfg.User = config.AppConfig.MySQL.Username
	cfg.Passwd = config.AppConfig.MySQL.Password
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%d", config.AppConfig.MySQL.Host, config.AppConfig.MySQL.Port)
	cfg.Params = map[string]string{
		"charset": "utf8mb4",
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal("Error connecting:", err)
	}
	defer db.Close()

	query := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		config.AppConfig.MySQL.Database,
	)
	if _, err := db.Exec(query); err != nil {
		log.Fatal("Error creating database:", err)
	}

	fmt.Printf("Database '%s' created successfully\n", config.AppConfig.MySQL.Database)
}
