package handlers

import (
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

	c.HTML(http.StatusOK, "partials/site-detail.html", gin.H{
		"site": site,
	})
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
	email := strings.TrimSpace(c.PostForm("email"))
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

	user, err := h.userService.Register(username, email, password)
	if err != nil {
		msg := "注册失败"
		if err == services.ErrUsernameExists {
			msg = "用户名已存在"
		} else if err == services.ErrEmailExists {
			msg = "邮箱已被注册"
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

func (h *APIHandler) ToggleFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}

	siteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的站点 ID"})
		return
	}

	isFav, _ := h.userService.IsFavorite(userID.(int64), siteID)
	if isFav {
		h.userService.RemoveFavorite(userID.(int64), siteID)
		c.JSON(http.StatusOK, gin.H{"is_fav": false})
	} else {
		h.userService.AddFavorite(userID.(int64), siteID)
		c.JSON(http.StatusOK, gin.H{"is_fav": true})
	}
}
