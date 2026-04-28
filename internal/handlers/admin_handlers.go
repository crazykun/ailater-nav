package handlers

import (
	"ai-later-nav/internal/models"
	"ai-later-nav/internal/services"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminPageData contains common fields for all admin pages
type AdminPageData struct {
	Username   string
	IsAdmin    bool
	PageTitle  string
	PageDesc   string
	Section    string
}

// AdminIndexData contains data specific to admin dashboard
type AdminIndexData struct {
	AdminPageData
	SiteCount int64
	UserCount int64
}

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
	siteCount, err := h.siteService.Count()
	if err != nil {
		log.Printf("Failed to get site count: %v", err)
		siteCount = 0
	}

	userCount, err := h.userService.CountUsers()
	if err != nil {
		log.Printf("Failed to get user count: %v", err)
		userCount = 0
	}

	data := AdminIndexData{
		AdminPageData: AdminPageData{
			Username:  c.GetString("username"),
			IsAdmin:   true,
			PageTitle: "仪表盘",
			PageDesc:  "查看后台关键指标和系统状态。",
			Section:   "dashboard",
		},
		SiteCount: siteCount,
		UserCount: userCount,
	}

	c.HTML(http.StatusOK, "admin-index.html", gin.H{
		"username":        data.Username,
		"siteCount":       data.SiteCount,
		"userCount":       data.UserCount,
		"adminSection":    data.Section,
		"pageTitle":       data.PageTitle,
		"pageDescription": data.PageDesc,
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
		"sites":           sites,
		"username":        c.GetString("username"),
		"adminSection":    "sites",
		"pageTitle":       "站点管理",
		"pageDescription": "集中维护站点资料、标签和推荐状态。",
	})
}

func (h *AdminHandler) AdminAddSiteForm(c *gin.Context) {
	categories, _ := h.siteService.GetCategories()
	c.HTML(http.StatusOK, "admin-add-site.html", gin.H{
		"categories":      categories,
		"username":        c.GetString("username"),
		"adminSection":    "sites",
		"pageTitle":       "添加站点",
		"pageDescription": "创建新的站点条目并设置分类、标签和推荐状态。",
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
		categories, _ := h.siteService.GetCategories()
		c.HTML(http.StatusOK, "admin-add-site.html", gin.H{
			"error":           "创建失败: " + err.Error(),
			"site":            site,
			"tagsString":      tagsStr,
			"categories":      categories,
			"username":        c.GetString("username"),
			"adminSection":    "sites",
			"pageTitle":       "添加站点",
			"pageDescription": "创建新的站点条目并设置分类、标签和推荐状态。",
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
		"site":            site,
		"tagsString":      strings.Join(site.Tags, ", "),
		"categories":      categories,
		"username":        c.GetString("username"),
		"adminSection":    "sites",
		"pageTitle":       "编辑站点",
		"pageDescription": "调整站点信息并维护推荐状态。",
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
		categories, _ := h.siteService.GetCategories()
		c.HTML(http.StatusOK, "admin-edit-site.html", gin.H{
			"error":           "更新失败: " + err.Error(),
			"site":            site,
			"tagsString":      tagsStr,
			"categories":      categories,
			"username":        c.GetString("username"),
			"adminSection":    "sites",
			"pageTitle":       "编辑站点",
			"pageDescription": "调整站点信息并维护推荐状态。",
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
		"users":           users,
		"username":        c.GetString("username"),
		"adminSection":    "users",
		"pageTitle":       "用户管理",
		"pageDescription": "查看注册用户和角色信息。",
	})
}

func (h *AdminHandler) AdminSettingsForm(c *gin.Context) {
	ss := services.GetSettingService()
	settings := ss.GetAllSettings()

	c.HTML(http.StatusOK, "admin-settings.html", gin.H{
		"settings":        settings,
		"username":        c.GetString("username"),
		"adminSection":    "settings",
		"pageTitle":       "系统设置",
		"pageDescription": "维护站点名称和版权信息。",
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
			"settings":        settings,
			"error":           "保存失败: " + err.Error(),
			"username":        c.GetString("username"),
			"adminSection":    "settings",
			"pageTitle":       "系统设置",
			"pageDescription": "维护站点名称和版权信息。",
		})
		return
	}

	settings := ss.GetAllSettings()
	c.HTML(http.StatusOK, "admin-settings.html", gin.H{
		"settings":        settings,
		"success":         "设置已保存",
		"username":        c.GetString("username"),
		"adminSection":    "settings",
		"pageTitle":       "系统设置",
		"pageDescription": "维护站点名称和版权信息。",
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
		"sites":           sitesWithStats,
		"username":        c.GetString("username"),
		"adminSection":    "stats",
		"pageTitle":       "访问统计",
		"pageDescription": "按站点查看 PV / UV 趋势概览。",
	})
}
