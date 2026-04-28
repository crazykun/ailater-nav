package handlers

import (
	"ai-later-nav/internal/models"
	"ai-later-nav/internal/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	siteService *services.SiteService
	userService *services.UserService
}

func NewAPIHandler() *APIHandler {
	return &APIHandler{
		siteService: services.NewSiteService(),
		userService: services.NewUserService(),
	}
}

func (h *APIHandler) SearchSites(c *gin.Context) {
	query := c.Query("q")
	category := c.Query("category")
	sortBy := c.DefaultQuery("sort", "rating")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	sites, total, err := h.siteService.Search(query, category, sortBy, page, 20)
	if err != nil {
		c.HTML(http.StatusOK, "partials/search-results.html", gin.H{
			"error": "搜索失败",
		})
		return
	}

	c.HTML(http.StatusOK, "partials/search-results.html", gin.H{
		"sites": sites,
		"total": total,
		"page":  page,
	})
}

func (h *APIHandler) SearchSuggestions(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.HTML(http.StatusOK, "partials/suggestions.html", gin.H{
			"suggestions": []string{},
		})
		return
	}

	suggestions, err := h.siteService.GetSearchSuggestions(query, 5)
	if err != nil {
		c.HTML(http.StatusOK, "partials/suggestions.html", gin.H{
			"suggestions": []string{},
		})
		return
	}

	c.HTML(http.StatusOK, "partials/suggestions.html", gin.H{
		"suggestions": suggestions,
		"query":       query,
	})
}

func (h *APIHandler) SiteDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.HTML(http.StatusOK, "partials/site-detail.html", gin.H{
			"error": "无效的站点 ID",
		})
		return
	}

	site, err := h.siteService.GetByID(id)
	if err != nil {
		c.HTML(http.StatusOK, "partials/site-detail.html", gin.H{
			"error": "加载失败",
		})
		return
	}

	ip := c.ClientIP()
	if err := h.siteService.IncrementVisits(id, ip); err != nil {
		// 记录日志但不影响响应
		// log.Printf("Failed to record visit: %v", err)
	}

	stats, err := h.siteService.GetSiteStats(id)
	if err != nil {
		stats = &models.SiteStats{}
	}

	c.HTML(http.StatusOK, "partials/site-detail.html", gin.H{
		"site":  site,
		"stats": stats,
	})
}

func (h *APIHandler) SiteStats(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点 ID"})
		return
	}

	stats, err := h.siteService.GetSiteStats(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *APIHandler) Login(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")

	if username == "" || password == "" {
		c.HTML(http.StatusOK, "partials/login-form.html", gin.H{
			"error": "请输入用户名和密码",
		})
		return
	}

	user, err := h.userService.Login(username, password)
	if err != nil {
		c.HTML(http.StatusOK, "partials/login-form.html", gin.H{
			"error": "用户名或密码错误",
		})
		return
	}

	token, err := generateToken(user)
	if err != nil {
		c.HTML(http.StatusOK, "partials/login-form.html", gin.H{
			"error": "登录失败，请重试",
		})
		return
	}

	c.SetCookie("token", token, 86400*7, "/", "", false, true)
	c.Header("HX-Redirect", "/")
	c.String(http.StatusOK, "<script>window.location.href='/'</script>")
}

func (h *APIHandler) Register(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	if username == "" || password == "" {
		c.HTML(http.StatusOK, "partials/register-form.html", gin.H{
			"error": "请输入用户名和密码",
		})
		return
	}

	if password != confirmPassword {
		c.HTML(http.StatusOK, "partials/register-form.html", gin.H{
			"error": "两次密码不一致",
		})
		return
	}

	if len(password) < 6 {
		c.HTML(http.StatusOK, "partials/register-form.html", gin.H{
			"error": "密码至少 6 位",
		})
		return
	}

	user, err := h.userService.Register(username, password)
	if err != nil {
		msg := "注册失败"
		if err == services.ErrUsernameExists {
			msg = "用户名已存在"
		}
		c.HTML(http.StatusOK, "partials/register-form.html", gin.H{
			"error": msg,
		})
		return
	}

	token, err := generateToken(user)
	if err != nil {
		c.HTML(http.StatusOK, "partials/register-form.html", gin.H{
			"error": "注册成功，请登录",
		})
		return
	}

	c.SetCookie("token", token, 86400*7, "/", "", false, true)
	c.Header("HX-Redirect", "/")
	c.String(http.StatusOK, "<script>window.location.href='/'</script>")
}

func (h *APIHandler) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.Header("HX-Redirect", "/")
	c.String(http.StatusOK, "<script>window.location.href='/'</script>")
}

func (h *APIHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.HTML(http.StatusOK, "partials/password-form.html", gin.H{
			"error": "请先登录",
		})
		return
	}

	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	if currentPassword == "" || newPassword == "" {
		c.HTML(http.StatusOK, "partials/password-form.html", gin.H{
			"error": "请填写当前密码和新密码",
		})
		return
	}

	if newPassword != confirmPassword {
		c.HTML(http.StatusOK, "partials/password-form.html", gin.H{
			"error": "两次新密码不一致",
		})
		return
	}

	if len(newPassword) < 6 {
		c.HTML(http.StatusOK, "partials/password-form.html", gin.H{
			"error": "新密码至少 6 位",
		})
		return
	}

	err := h.userService.ChangePassword(userID.(int64), currentPassword, newPassword)
	if err != nil {
		msg := "修改失败"
		if err == services.ErrInvalidPassword {
			msg = "当前密码错误"
		}
		c.HTML(http.StatusOK, "partials/password-form.html", gin.H{
			"error": msg,
		})
		return
	}

	c.HTML(http.StatusOK, "partials/password-form.html", gin.H{
		"success": "密码修改成功",
	})
}

func (h *APIHandler) Setup(c *gin.Context) {
	hasUser, _ := h.userService.HasAnyUser()
	if hasUser {
		c.Header("HX-Redirect", "/")
		c.String(http.StatusOK, "<script>window.location.href='/'</script>")
		return
	}

	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")

	if username == "" || password == "" {
		c.HTML(http.StatusOK, "partials/setup-form.html", gin.H{
			"error": "请输入用户名和密码",
		})
		return
	}

	if password != confirmPassword {
		c.HTML(http.StatusOK, "partials/setup-form.html", gin.H{
			"error": "两次密码不一致",
		})
		return
	}

	if len(password) < 6 {
		c.HTML(http.StatusOK, "partials/setup-form.html", gin.H{
			"error": "密码至少 6 位",
		})
		return
	}

	user, err := h.userService.RegisterAdmin(username, password)
	if err != nil {
		c.HTML(http.StatusOK, "partials/setup-form.html", gin.H{
			"error": "创建失败: " + err.Error(),
		})
		return
	}

	services.GetSettingService().SeedDefaults()

	token, err := generateToken(user)
	if err != nil {
		c.HTML(http.StatusOK, "partials/setup-form.html", gin.H{
			"error": "创建成功，请登录",
		})
		return
	}

	c.SetCookie("token", token, 86400*7, "/", "", false, true)
	c.Header("HX-Redirect", "/admin/")
	c.String(http.StatusOK, "<script>window.location.href='/admin/'</script>")
}

func (h *APIHandler) ToggleFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.Header("HX-Redirect", "/login")
		c.Status(http.StatusUnauthorized)
		return
	}

	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	isFav, _ := h.userService.IsFavorite(userID.(int64), siteID)
	if isFav {
		h.userService.RemoveFavorite(userID.(int64), siteID)
	} else {
		h.userService.AddFavorite(userID.(int64), siteID)
	}

	from := c.Query("from")
	if from == "dashboard" {
		favoriteIDs, _ := h.userService.GetFavoriteIDs(userID.(int64))
		var favoriteSites []models.SiteDisplay
		if len(favoriteIDs) > 0 {
			favoriteSites, _ = h.siteService.GetByIDs(favoriteIDs)
		}
		c.HTML(http.StatusOK, "partials/favorites-list.html", gin.H{
			"favoriteSites": favoriteSites,
		})
		return
	}

	newIsFav := !isFav
	c.HTML(http.StatusOK, "partials/favorite-btn.html", gin.H{
		"siteID": siteID,
		"isFav":  newIsFav,
	})
}
