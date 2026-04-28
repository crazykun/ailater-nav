package web

import (
	"html/template"
	"io/fs"
	"strings"
)

var funcMap = template.FuncMap{
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
}

func BuildPageTemplates(root fs.FS) map[string]*template.Template {
	return map[string]*template.Template{
		"index.html":     template.Must(template.New("index.html").Funcs(funcMap).ParseFS(root, "layout.html", "index.html", "partials/*.html")),
		"login.html":     template.Must(template.New("login.html").Funcs(funcMap).ParseFS(root, "layout.html", "login.html", "partials/*.html")),
		"register.html":  template.Must(template.New("register.html").Funcs(funcMap).ParseFS(root, "layout.html", "register.html", "partials/*.html")),
		"dashboard.html": template.Must(template.New("dashboard.html").Funcs(funcMap).ParseFS(root, "layout.html", "dashboard.html", "partials/*.html")),
		"setup.html":     template.Must(template.New("setup.html").Funcs(funcMap).ParseFS(root, "layout.html", "setup.html", "partials/*.html")),
	}
}
