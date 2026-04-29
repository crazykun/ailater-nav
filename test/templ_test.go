package test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"ai-later-nav/internal/web"
)

func TestIndexTemplateDoesNotRenderRegisterContent(t *testing.T) {
	tmpl := web.BuildPageTemplates(os.DirFS("../templates"))["index.html"]

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "index.html", map[string]any{
		"sites":      []any{},
		"isLoggedIn": false,
		"username":   "",
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	if strings.Contains(html, "创建新账号") {
		t.Fatalf("index.html rendered register content: %s", html)
	}
	if !strings.Contains(html, "AI 工具导航") {
		t.Fatalf("index.html did not render homepage content: %s", html)
	}
}

func TestLoginTemplateDoesNotRenderRegisterContent(t *testing.T) {
	tmpl := web.BuildPageTemplates(os.DirFS("../templates"))["login.html"]

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "login.html", map[string]any{})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	if strings.Contains(html, "创建新账号") {
		t.Fatalf("login.html rendered register content: %s", html)
	}
	if !strings.Contains(html, "欢迎回来") {
		t.Fatalf("login.html did not render login content: %s", html)
	}
}

func TestAdminIndexTemplateRendersFromSharedTemplateSet(t *testing.T) {
	tmpl := web.BuildSharedTemplates(os.DirFS("../templates"))

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "admin-index.html", map[string]any{
		"username":     "admin",
		"siteCount":    12,
		"userCount":    3,
		"adminSection": "dashboard",
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	if !strings.Contains(html, "仪表盘") {
		t.Fatalf("admin-index.html did not render dashboard content: %s", html)
	}
}

func TestAdminIndexTemplateUsesSharedShell(t *testing.T) {
	tmpl := web.BuildSharedTemplates(os.DirFS("../templates"))

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "admin-index.html", map[string]any{
		"username":     "admin",
		"siteCount":    12,
		"userCount":    3,
		"adminSection": "dashboard",
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	for _, needle := range []string{"/static/css/style.css", "admin-shell", "admin-sidebar", "admin-topbar", "admin-stat-card"} {
		if !strings.Contains(html, needle) {
			t.Fatalf("admin-index.html missing %q in rendered output: %s", needle, html)
		}
	}
}

func TestAdminSitesTemplateUsesResponsiveTableShell(t *testing.T) {
	tmpl := web.BuildSharedTemplates(os.DirFS("../templates"))

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "admin-sites.html", map[string]any{
		"username":        "admin",
		"adminSection":    "sites",
		"pageTitle":       "站点管理",
		"pageDescription": "管理站点内容。",
		"sites":           []any{},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	for _, needle := range []string{"overflow-x-auto", "admin-table", "admin-btn-primary"} {
		if !strings.Contains(html, needle) {
			t.Fatalf("admin-sites.html missing %q in rendered output: %s", needle, html)
		}
	}
}

func TestAdminUsersTemplateUsesSharedShell(t *testing.T) {
	tmpl := web.BuildSharedTemplates(os.DirFS("../templates"))

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "admin-users.html", map[string]any{
		"username":        "admin",
		"adminSection":    "users",
		"pageTitle":       "用户管理",
		"pageDescription": "查看用户账号信息。",
		"users":           []any{},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	if !strings.Contains(html, "admin-table") {
		t.Fatalf("admin-users.html did not render shared table shell: %s", html)
	}
}

func TestAdminStatsTemplateUsesSharedShell(t *testing.T) {
	tmpl := web.BuildSharedTemplates(os.DirFS("../templates"))

	// Create test data with a site that has stats
	type SiteWithStats struct {
		Name     string
		Category string
		Logo     string
		Color    string
		Initials string
		Stats    *struct {
			PV      int64
			UV      int64
			TodayPV int64
			TodayUV int64
			WeekPV  int64
			WeekUV  int64
		}
	}

	testSites := []SiteWithStats{
		{
			Name:     "Test Site",
			Category: "AI",
			Logo:     "",
			Color:    "#667eea",
			Initials: "TS",
			Stats: &struct {
				PV      int64
				UV      int64
				TodayPV int64
				TodayUV int64
				WeekPV  int64
				WeekUV  int64
			}{
				PV:      100,
				UV:      50,
				TodayPV: 10,
				TodayUV: 5,
				WeekPV:  70,
				WeekUV:  35,
			},
		},
	}

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "admin-stats.html", map[string]any{
		"username":        "admin",
		"adminSection":    "stats",
		"pageTitle":       "访问统计",
		"pageDescription": "按站点查看 PV / UV 趋势概览。",
		"sites":           testSites,
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	for _, needle := range []string{"admin-table", "admin-metric-number"} {
		if !strings.Contains(html, needle) {
			t.Fatalf("admin-stats.html missing %q in rendered output: %s", needle, html)
		}
	}
}

func TestAdminSettingsTemplateUsesSharedFormAndFlash(t *testing.T) {
	tmpl := web.BuildSharedTemplates(os.DirFS("../templates"))

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "admin-settings.html", map[string]any{
		"username":        "admin",
		"adminSection":    "settings",
		"pageTitle":       "系统设置",
		"pageDescription": "维护站点基础配置。",
		"settings":        map[string]string{"site_name": "AI Later"},
		"success":         "设置已保存",
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	for _, needle := range []string{"admin-alert-success", "admin-form-grid", "admin-input", "保存设置"} {
		if !strings.Contains(html, needle) {
			t.Fatalf("admin-settings.html missing %q in rendered output: %s", needle, html)
		}
	}
}

func TestIndexTemplateRendersDisplayTagClasses(t *testing.T) {
	tmpl := web.BuildPageTemplates(os.DirFS("../templates"))["index.html"]

	type DisplayTag struct {
		Name  string
		Class string
	}
	type SiteView struct {
		ID          int64
		Name        string
		Description string
		URL         string
		Logo        string
		Color       string
		Initials    string
		Tags        []string
		DisplayTags []DisplayTag
		TodayUV     int64
		IsFav       bool
	}

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "index.html", map[string]any{
		"sites": []SiteView{{
			ID:          1,
			Name:        "Test Site",
			Description: "desc",
			URL:         "https://example.com",
			Color:       "#123456",
			Initials:    "TS",
			DisplayTags: []DisplayTag{{
				Name:  "AI对话",
				Class: "bg-emerald-50 text-emerald-700 border border-emerald-200",
			}},
		}},
		"isLoggedIn": false,
		"username":   "",
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	if !strings.Contains(html, "bg-emerald-50 text-emerald-700 border border-emerald-200") {
		t.Fatalf("index.html did not render display tag class: %s", html)
	}
	if strings.Contains(html, "class=\"tag-badge ") {
		t.Fatalf("index.html still rendered tag-badge class that overrides stable colors: %s", html)
	}
	if strings.Contains(html, "bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 border border-blue-100 dark:border-blue-800") {
		t.Fatalf("index.html still rendered hard-coded blue tag class: %s", html)
	}
}

func TestSearchResultsTemplateRendersDisplayTagClasses(t *testing.T) {
	tmpl := web.BuildSharedTemplates(os.DirFS("../templates"))

	type DisplayTag struct {
		Name  string
		Class string
	}
	type SiteView struct {
		ID          int64
		Name        string
		Description string
		URL         string
		Logo        string
		Color       string
		Initials    string
		Tags        []string
		DisplayTags []DisplayTag
		TodayUV     int64
		IsFav       bool
	}

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "partials/search-results.html", map[string]any{
		"sites": []SiteView{{
			ID:          1,
			Name:        "Test Site",
			Description: "desc",
			URL:         "https://example.com",
			Color:       "#123456",
			Initials:    "TS",
			DisplayTags: []DisplayTag{{Name: "AI对话", Class: "bg-rose-50 text-rose-700 border border-rose-200"}},
		}},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	if !strings.Contains(html, "bg-rose-50 text-rose-700 border border-rose-200") {
		t.Fatalf("search-results did not render display tag class: %s", html)
	}
	if strings.Contains(html, "bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400 border border-blue-100 dark:border-blue-800") {
		t.Fatalf("search-results still rendered hard-coded blue tag class: %s", html)
	}
}

func TestSiteDetailTemplateRendersDisplayTagClasses(t *testing.T) {
	tmpl := web.BuildSharedTemplates(os.DirFS("../templates"))

	type DisplayTag struct {
		Name  string
		Class string
	}
	type SiteView struct {
		Name        string
		Category    string
		Description string
		Rating      float64
		Visits      int64
		URL         string
		Logo        string
		Color       string
		Initials    string
		Tags        []string
		DisplayTags []DisplayTag
	}

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "partials/site-detail.html", map[string]any{
		"site": SiteView{
			Name:        "Test Site",
			Category:    "AI",
			Description: "desc",
			Rating:      4.5,
			Visits:      10,
			DisplayTags: []DisplayTag{{Name: "AI对话", Class: "bg-violet-50 text-violet-700 border border-violet-200"}},
		},
		"stats": map[string]any{},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	if !strings.Contains(html, "bg-violet-50 text-violet-700 border border-violet-200") {
		t.Fatalf("site-detail did not render display tag class: %s", html)
	}
	if strings.Contains(html, "px-3 py-1 bg-blue-50 text-blue-600 rounded-full text-sm") {
		t.Fatalf("site-detail still rendered hard-coded blue tag class: %s", html)
	}
}

func TestAdminSitesTemplateRendersDisplayTagClasses(t *testing.T) {
	tmpl := web.BuildSharedTemplates(os.DirFS("../templates"))

	type DisplayTag struct {
		Name  string
		Class string
	}
	type SiteView struct {
		ID          int64
		Name        string
		Description string
		URL         string
		Logo        string
		Color       string
		Initials    string
		Tags        []string
		DisplayTags []DisplayTag
	}

	var out bytes.Buffer
	err := tmpl.ExecuteTemplate(&out, "admin-sites.html", map[string]any{
		"username":        "admin",
		"adminSection":    "sites",
		"pageTitle":       "站点管理",
		"pageDescription": "管理站点内容。",
		"sites": []SiteView{{
			ID:          1,
			Name:        "Test Site",
			Description: "desc",
			URL:         "https://example.com",
			Color:       "#123456",
			Initials:    "TS",
			DisplayTags: []DisplayTag{{Name: "AI对话", Class: "bg-sky-50 text-sky-700 border border-sky-200"}},
		}},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	html := out.String()
	if !strings.Contains(html, "bg-sky-50 text-sky-700 border border-sky-200") {
		t.Fatalf("admin-sites did not render display tag class: %s", html)
	}
	if strings.Contains(html, "class=\"admin-badge\"") {
		t.Fatalf("admin-sites still rendered old admin badge markup: %s", html)
	}
}
