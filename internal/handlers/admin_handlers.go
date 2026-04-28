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
	SiteName   string
}

// getCommonAdminData returns common template data for admin pages
func getCommonAdminData(c *gin.Context, section, title, description string) gin.H {
	return gin.H{
		"username":        c.GetString("username"),
		"adminSection":    section,
		"pageTitle":       title,
		"pageDescription": description,
		"SiteName":        c.GetString("SiteName"),
	}
}

// AdminIndexData contains data specific to admin dashboard
type AdminIndexData struct {
	AdminPageData
	SiteCount    int64
	UserCount    int64
	TodayUsers   int64
	DashboardStats *models.DashboardStats
	TopSites     []models.Site
	RecentUsers  []*models.User
	CategoryStats []models.CategoryStat
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

	todayUsers, err := h.userService.CountTodayUsers()
	if err != nil {
		log.Printf("Failed to get today users: %v", err)
		todayUsers = 0
	}

	dashboardStats, err := h.siteService.GetDashboardStats()
	if err != nil {
		log.Printf("Failed to get dashboard stats: %v", err)
		dashboardStats = &models.DashboardStats{}
	}

	topSites, err := h.siteService.GetTopSites(5)
	if err != nil {
		log.Printf("Failed to get top sites: %v", err)
		topSites = []models.Site{}
	}

	recentUsers, err := h.userService.GetRecentUsers(5)
	if err != nil {
		log.Printf("Failed to get recent users: %v", err)
		recentUsers = []*models.User{}
	}

	categoryStats, err := h.siteService.GetCategoryStats()
	if err != nil {
		log.Printf("Failed to get category stats: %v", err)
		categoryStats = []models.CategoryStat{}
	}

	common := getCommonAdminData(c, "dashboard", "仪表盘", "查看后台关键指标和系统状态。")
	common["siteCount"] = siteCount
	common["userCount"] = userCount
	common["todayUsers"] = todayUsers
	common["dashboardStats"] = dashboardStats
	common["topSites"] = topSites
	common["recentUsers"] = recentUsers
	common["categoryStats"] = categoryStats
	c.HTML(http.StatusOK, "admin-index.html", common)
}

func (h *AdminHandler) AdminSites(c *gin.Context) {
	sites, err := h.siteService.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "加载站点失败",
		})
		return
	}

	for i, j := 0, len(sites)-1; i < j; i, j = i+1, j-1 {
		sites[i], sites[j] = sites[j], sites[i]
	}

	common := getCommonAdminData(c, "sites", "站点管理", "集中维护站点资料、标签和推荐状态。")
	common["sites"] = sites
	c.HTML(http.StatusOK, "admin-sites.html", common)
}

func (h *AdminHandler) AdminAddSiteForm(c *gin.Context) {
	categories, _ := h.siteService.GetCategories()
	common := getCommonAdminData(c, "sites", "添加站点", "创建新的站点条目并设置分类、标签和推荐状态。")
	common["categories"] = categories
	c.HTML(http.StatusOK, "admin-add-site.html", common)
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
		common := getCommonAdminData(c, "sites", "添加站点", "创建新的站点条目并设置分类、标签和推荐状态。")
		common["error"] = "创建失败: " + err.Error()
		common["site"] = site
		common["tagsString"] = tagsStr
		common["categories"] = categories
		c.HTML(http.StatusOK, "admin-add-site.html", common)
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
	common := getCommonAdminData(c, "sites", "编辑站点", "调整站点信息并维护推荐状态。")
	common["site"] = site
	common["tagsString"] = strings.Join(site.Tags, ", ")
	common["categories"] = categories
	c.HTML(http.StatusOK, "admin-edit-site.html", common)
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
		common := getCommonAdminData(c, "sites", "编辑站点", "调整站点信息并维护推荐状态。")
		common["error"] = "更新失败: " + err.Error()
		common["site"] = site
		common["tagsString"] = tagsStr
		common["categories"] = categories
		c.HTML(http.StatusOK, "admin-edit-site.html", common)
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

	common := getCommonAdminData(c, "users", "用户管理", "查看注册用户和角色信息。")
	common["users"] = users
	c.HTML(http.StatusOK, "admin-users.html", common)
}

func (h *AdminHandler) AdminSettingsForm(c *gin.Context) {
	ss := services.GetSettingService()
	settings := ss.GetAllSettings()

	common := getCommonAdminData(c, "settings", "系统设置", "维护站点名称和版权信息。")
	common["settings"] = settings
	c.HTML(http.StatusOK, "admin-settings.html", common)
}

func (h *AdminHandler) AdminSettingsSave(c *gin.Context) {
	ss := services.GetSettingService()
	updates := map[string]string{
		"site_name": strings.TrimSpace(c.PostForm("site_name")),
		"copyright": strings.TrimSpace(c.PostForm("copyright")),
	}

	if err := ss.UpdateMultiple(updates); err != nil {
		settings := ss.GetAllSettings()
		common := getCommonAdminData(c, "settings", "系统设置", "维护站点名称和版权信息。")
		common["settings"] = settings
		common["error"] = "保存失败: " + err.Error()
		c.HTML(http.StatusOK, "admin-settings.html", common)
		return
	}

	settings := ss.GetAllSettings()
	common := getCommonAdminData(c, "settings", "系统设置", "维护站点名称和版权信息。")
	common["settings"] = settings
	common["success"] = "设置已保存"
	c.HTML(http.StatusOK, "admin-settings.html", common)
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

	common := getCommonAdminData(c, "stats", "访问统计", "按站点查看 PV / UV 趋势概览。")
	common["sites"] = sitesWithStats
	c.HTML(http.StatusOK, "admin-stats.html", common)
}
