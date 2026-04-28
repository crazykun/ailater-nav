package handlers

import (
	"ai-later-nav/internal/models"
	"ai-later-nav/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	siteService *services.SiteService
	userService *services.UserService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		siteService: services.NewSiteService(),
		userService: services.NewUserService(),
	}
}

func (h *AdminHandler) AdminIndex(c *gin.Context) {
	siteCount, _ := h.siteService.Count()
	userCount, _ := h.userService.CountUsers()
	copyright, _ := c.Get("Copyright")
	c.HTML(http.StatusOK, "admin-index.html", gin.H{
		"Copyright": copyright,
		"isAdmin":   true,
		"username":  c.GetString("username"),
		"siteCount": siteCount,
		"userCount": userCount,
	})
}

func (h *AdminHandler) AdminSites(c *gin.Context) {
	sites, err := h.siteService.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "加载站点失败",
		})
		return
	}

	c.HTML(http.StatusOK, "admin-sites.html", gin.H{
		"sites":    sites,
		"isAdmin":  true,
		"username": c.GetString("username"),
	})
}

func (h *AdminHandler) AdminAddSiteForm(c *gin.Context) {
	categories, _ := h.siteService.GetCategories()
	c.HTML(http.StatusOK, "admin-add-site.html", gin.H{
		"categories": categories,
		"isAdmin":    true,
		"username":   c.GetString("username"),
	})
}

func (h *AdminHandler) AdminAddSite(c *gin.Context) {
	site := &models.Site{
		Name:        strings.TrimSpace(c.PostForm("Name")),
		URL:         strings.TrimSpace(c.PostForm("URL")),
		Description: strings.TrimSpace(c.PostForm("Description")),
		Logo:        strings.TrimSpace(c.PostForm("Logo")),
		Category:    strings.TrimSpace(c.PostForm("Category")),
	}

	if ratingStr := c.PostForm("Rating"); ratingStr != "" {
		if rating, err := strconv.ParseFloat(ratingStr, 64); err == nil {
			site.Rating = rating
		}
	}

	site.Featured = c.PostForm("Featured") == "on"

	tagsStr := c.PostForm("Tags")
	var tags []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	if _, err := h.siteService.Create(site, tags); err != nil {
		c.HTML(http.StatusOK, "admin-add-site.html", gin.H{
			"error":      "创建失败: " + err.Error(),
			"site":       site,
			"tagsString": tagsStr,
			"isAdmin":    true,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/sites")
}

func (h *AdminHandler) AdminEditSiteForm(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "无效的站点 ID",
		})
		return
	}

	site, err := h.siteService.GetByID(id)
	if err != nil || site == nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "站点不存在",
		})
		return
	}

	categories, _ := h.siteService.GetCategories()
	c.HTML(http.StatusOK, "admin-edit-site.html", gin.H{
		"site":       site,
		"tagsString": strings.Join(site.Tags, ", "),
		"categories": categories,
		"isAdmin":    true,
		"username":   c.GetString("username"),
	})
}

func (h *AdminHandler) AdminEditSite(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "无效的站点 ID",
		})
		return
	}

	site := &models.Site{
		ID:          id,
		Name:        strings.TrimSpace(c.PostForm("Name")),
		URL:         strings.TrimSpace(c.PostForm("URL")),
		Description: strings.TrimSpace(c.PostForm("Description")),
		Logo:        strings.TrimSpace(c.PostForm("Logo")),
		Category:    strings.TrimSpace(c.PostForm("Category")),
	}

	if ratingStr := c.PostForm("Rating"); ratingStr != "" {
		if rating, err := strconv.ParseFloat(ratingStr, 64); err == nil {
			site.Rating = rating
		}
	}

	site.Featured = c.PostForm("Featured") == "on"

	tagsStr := c.PostForm("Tags")
	var tags []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	if err := h.siteService.Update(site, tags); err != nil {
		c.HTML(http.StatusOK, "admin-edit-site.html", gin.H{
			"error":      "更新失败: " + err.Error(),
			"site":       site,
			"tagsString": tagsStr,
			"isAdmin":    true,
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin/sites")
}

func (h *AdminHandler) AdminDeleteSite(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点 ID"})
		return
	}

	if err := h.siteService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.Redirect(http.StatusFound, "/admin/sites")
}

func (h *AdminHandler) AdminUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "加载用户失败",
		})
		return
	}

	c.HTML(http.StatusOK, "admin-users.html", gin.H{
		"users":    users,
		"isAdmin":  true,
		"username": c.GetString("username"),
	})
}

func (h *AdminHandler) AdminSettingsForm(c *gin.Context) {
	ss := services.GetSettingService()
	settings := ss.GetAllSettings()

	c.HTML(http.StatusOK, "admin-settings.html", gin.H{
		"settings": settings,
		"isAdmin":  true,
		"username": c.GetString("username"),
	})
}

func (h *AdminHandler) AdminSettingsSave(c *gin.Context) {
	ss := services.GetSettingService()
	updates := map[string]string{
		"site_name": strings.TrimSpace(c.PostForm("site_name")),
		"copyright": strings.TrimSpace(c.PostForm("copyright")),
	}

	if err := ss.UpdateMultiple(updates); err != nil {
		settings := ss.GetAllSettings()
		c.HTML(http.StatusOK, "admin-settings.html", gin.H{
			"settings": settings,
			"error":    "保存失败: " + err.Error(),
			"isAdmin":  true,
			"username": c.GetString("username"),
		})
		return
	}

	settings := ss.GetAllSettings()
	c.HTML(http.StatusOK, "admin-settings.html", gin.H{
		"settings": settings,
		"success":  "设置已保存",
		"isAdmin":  true,
		"username": c.GetString("username"),
	})
}

func (h *AdminHandler) AdminStats(c *gin.Context) {
	sites, err := h.siteService.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "加载站点失败",
		})
		return
	}

	allStats, err := h.siteService.GetAllSitesStats()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "加载统计失败",
		})
		return
	}

	statsMap := make(map[int64]models.SiteStats)
	for _, s := range allStats {
		statsMap[s.SiteID] = s
	}

	type SiteWithStats struct {
		models.SiteDisplay
		Stats *models.SiteStats
	}

	var sitesWithStats []SiteWithStats
	for _, site := range sites {
		sws := SiteWithStats{SiteDisplay: site}
		if stats, ok := statsMap[site.ID]; ok {
			sws.Stats = &stats
		}
		sitesWithStats = append(sitesWithStats, sws)
	}

	c.HTML(http.StatusOK, "admin-stats.html", gin.H{
		"sites":    sitesWithStats,
		"isAdmin":  true,
		"username": c.GetString("username"),
	})
}
