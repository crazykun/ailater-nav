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
