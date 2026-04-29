package handlers

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"ai-later-nav/internal/models"
	"ai-later-nav/internal/services"
	"github.com/gin-gonic/gin"
)

type PageHandler struct {
	siteService   *services.SiteService
	pageTemplates map[string]*template.Template
}

func NewPageHandler(pageTemplates map[string]*template.Template) *PageHandler {
	return &PageHandler{
		siteService:   services.NewSiteService(),
		pageTemplates: pageTemplates,
	}
}

func (h *PageHandler) renderPage(c *gin.Context, name string, data gin.H) error {
	tmpl, ok := h.pageTemplates[name]
	if !ok {
		return errors.New("page template not found: " + name)
	}

	if data == nil {
		data = gin.H{}
	}
	if v, exists := c.Get("Copyright"); exists {
		data["Copyright"] = v
	}
	if v, exists := c.Get("SiteName"); exists {
		data["SiteName"] = v
	}
	if v, exists := c.Get("isLoggedIn"); exists {
		data["isLoggedIn"] = v
	}
	if v, exists := c.Get("username"); exists {
		data["username"] = v
	}
	if v, exists := c.Get("role"); exists {
		data["role"] = v
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
	return nil
}

func (h *PageHandler) HomePage(c *gin.Context) {
	sites, err := h.siteService.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "加载站点数据失败",
		})
		return
	}

	if c.GetBool("isLoggedIn") {
		userService := services.NewUserService()
		if favIDs, err := userService.GetFavoriteIDs(c.GetInt64("user_id")); err == nil && len(favIDs) > 0 {
			favSet := make(map[int64]bool, len(favIDs))
			for _, id := range favIDs {
				favSet[id] = true
			}
			for i := range sites {
				sites[i].IsFav = favSet[sites[i].ID]
			}
		}
	}

	categories, _ := h.siteService.GetCategories()
	copyright, _ := c.Get("Copyright")

	if err := h.renderPage(c, "index.html", gin.H{
		"sites":      sites,
		"categories": categories,
		"Copyright":  copyright,
		"isLoggedIn": c.GetBool("isLoggedIn"),
		"username":   c.GetString("username"),
	}); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "页面渲染失败",
		})
	}
}

func (h *PageHandler) SearchPage(c *gin.Context) {
	query := c.Query("q")
	category := c.Query("category")
	sortBy := c.DefaultQuery("sort", "rating")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	sites, total, err := h.siteService.Search(query, category, sortBy, page, 20)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "搜索失败",
		})
		return
	}

	if c.GetBool("isLoggedIn") {
		userService := services.NewUserService()
		if favIDs, err := userService.GetFavoriteIDs(c.GetInt64("user_id")); err == nil && len(favIDs) > 0 {
			favSet := make(map[int64]bool, len(favIDs))
			for _, id := range favIDs {
				favSet[id] = true
			}
			for i := range sites {
				sites[i].IsFav = favSet[sites[i].ID]
			}
		}
	}

	categories, _ := h.siteService.GetCategories()
	copyright, _ := c.Get("Copyright")

	if err := h.renderPage(c, "index.html", gin.H{
		"sites":            sites,
		"categories":       categories,
		"query":            query,
		"selectedCategory": category,
		"selectedSort":     sortBy,
		"total":            total,
		"page":             page,
		"Copyright":        copyright,
		"isLoggedIn":       c.GetBool("isLoggedIn"),
		"username":         c.GetString("username"),
	}); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "页面渲染失败",
		})
	}
}

func (h *PageHandler) LoginPage(c *gin.Context) {
	if err := h.renderPage(c, "login.html", gin.H{
		"Copyright": c.GetString("Copyright"),
	}); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "页面渲染失败",
		})
	}
}

func (h *PageHandler) RegisterPage(c *gin.Context) {
	if err := h.renderPage(c, "register.html", gin.H{
		"Copyright": c.GetString("Copyright"),
	}); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "页面渲染失败",
		})
	}
}

func (h *PageHandler) UserDashboard(c *gin.Context) {
	userID := c.GetInt64("user_id")
	userService := services.NewUserService()

	favoriteIDs, _ := userService.GetFavoriteIDs(userID)

	var favoriteSites []models.SiteDisplay
	if len(favoriteIDs) > 0 {
		favoriteSites, _ = h.siteService.GetByIDs(favoriteIDs)
	}

	copyright, _ := c.Get("Copyright")
	if err := h.renderPage(c, "dashboard.html", gin.H{
		"favoriteIDs":   favoriteIDs,
		"favoriteSites": favoriteSites,
		"Copyright":     copyright,
		"isLoggedIn":    true,
		"username":      c.GetString("username"),
	}); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "页面渲染失败",
		})
	}
}

func (h *PageHandler) SetupPage(c *gin.Context) {
	userService := services.NewUserService()
	hasUser, _ := userService.HasAnyUser()
	if hasUser {
		c.Redirect(http.StatusFound, "/")
		return
	}

	copyright, _ := c.Get("Copyright")
	if err := h.renderPage(c, "setup.html", gin.H{
		"Copyright": copyright,
	}); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "页面渲染失败",
		})
	}
}
