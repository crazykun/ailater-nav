package main

import (
	"ai-later-nav/internal/config"
	"ai-later-nav/internal/database"
	"ai-later-nav/internal/handlers"
	"ai-later-nav/internal/middleware"
	"ai-later-nav/internal/web"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

func main() {
	if err := config.LoadConfig(); err != nil {
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

	r := gin.Default()

	tmplSub, _ := fs.Sub(templateFS, "templates")
	pageTemplates := web.BuildPageTemplates(tmplSub)
	sharedTemplates := template.Must(template.ParseFS(tmplSub, "error.html", "admin/*.html", "partials/*.html"))
	r.SetHTMLTemplate(sharedTemplates)

	staticSub, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/static", http.FS(staticSub))

	r.Use(middleware.OptionalAuth())
	r.Use(middleware.AddGlobalContext())

	pageHandler := handlers.NewPageHandler(pageTemplates)
	apiHandler := handlers.NewAPIHandler()
	adminHandler := handlers.NewAdminHandler()

	r.GET("/", pageHandler.HomePage)
	r.GET("/search", pageHandler.SearchPage)

	r.GET("/login", pageHandler.LoginPage)
	r.GET("/register", pageHandler.RegisterPage)

	r.POST("/api/auth/login", apiHandler.Login)
	r.POST("/api/auth/register", apiHandler.Register)
	r.POST("/api/auth/logout", apiHandler.Logout)

	r.GET("/api/search", apiHandler.SearchSites)
	r.GET("/api/sites/:id", apiHandler.SiteDetail)

	auth := r.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/dashboard", pageHandler.UserDashboard)
		auth.POST("/api/favorites/:id", apiHandler.ToggleFavorite)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminRequired())
	{
		admin.GET("/", adminHandler.AdminIndex)
		admin.GET("/sites", adminHandler.AdminSites)
		admin.GET("/sites/add", adminHandler.AdminAddSiteForm)
		admin.POST("/sites/add", adminHandler.AdminAddSite)
		admin.GET("/sites/edit/:id", adminHandler.AdminEditSiteForm)
		admin.POST("/sites/edit/:id", adminHandler.AdminEditSite)
		admin.GET("/sites/delete/:id", adminHandler.AdminDeleteSite)
	}

	port := config.AppConfig.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on http://localhost:%s", port)
	r.Run(":" + port)
}
