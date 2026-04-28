package main

import (
	"ai-later-nav/internal/config"
	"ai-later-nav/internal/database"
	"ai-later-nav/internal/database/repository"
	"ai-later-nav/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	if err := config.LoadRequiredConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.Init(database.MySQLConfig{
		Host:     config.AppConfig.MySQL.Host,
		Port:     config.AppConfig.MySQL.Port,
		Username: config.AppConfig.MySQL.Username,
		Password: config.AppConfig.MySQL.Password,
		Database: config.AppConfig.MySQL.Database,
	}); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()

	if err := database.RunMigrations(database.DB, "internal/database/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := migrateSites(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}

type OldSite struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Description string   `json:"description"`
	Logo        string   `json:"logo"`
	Tags        []string `json:"tags"`
	Category    string   `json:"category"`
	Rating      float64  `json:"rating"`
	Featured    bool     `json:"featured"`
}

func migrateSites() error {
	data, err := os.ReadFile("./data/ai.json")
	if err != nil {
		return fmt.Errorf("read ai.json: %w", err)
	}

	var oldSites []OldSite
	if err := json.Unmarshal(data, &oldSites); err != nil {
		return fmt.Errorf("parse ai.json: %w", err)
	}

	siteRepo := repository.NewSiteRepository()

	var created, skipped, failed int
	for _, old := range oldSites {
		existing, err := siteRepo.GetByName(old.Name)
		if err != nil {
			log.Printf("Failed to check existing site %s: %v", old.Name, err)
			failed++
			continue
		}
		if existing != nil {
			log.Printf("Skipping existing site: %s", old.Name)
			skipped++
			continue
		}

		site := &models.Site{
			Name:        old.Name,
			URL:         old.URL,
			Description: old.Description,
			Logo:        old.Logo,
			Category:    old.Category,
			Rating:      old.Rating,
			Featured:    old.Featured,
		}

		id, err := siteRepo.Create(site)
		if err != nil {
			log.Printf("Failed to create site %s: %v", old.Name, err)
			failed++
			continue
		}

		if len(old.Tags) > 0 {
			if err := siteRepo.SetTags(id, old.Tags); err != nil {
				log.Printf("Failed to set tags for %s: %v", old.Name, err)
				failed++
				continue
			}
		}

		log.Printf("Created site: %s (ID: %d)", old.Name, id)
		created++
	}

	log.Printf("Migration summary: %d created, %d skipped, %d failed", created, skipped, failed)
	if failed > 0 {
		return fmt.Errorf("migration completed with %d failed records", failed)
	}
	return nil
}
