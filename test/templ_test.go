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
