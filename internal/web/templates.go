package web

import (
	"html/template"
	"io/fs"
)

func BuildPageTemplates(root fs.FS) map[string]*template.Template {
	return map[string]*template.Template{
		"index.html":     template.Must(template.ParseFS(root, "layout.html", "index.html", "partials/*.html")),
		"login.html":     template.Must(template.ParseFS(root, "layout.html", "login.html", "partials/*.html")),
		"register.html":  template.Must(template.ParseFS(root, "layout.html", "register.html", "partials/*.html")),
		"dashboard.html": template.Must(template.ParseFS(root, "layout.html", "dashboard.html", "partials/*.html")),
	}
}
