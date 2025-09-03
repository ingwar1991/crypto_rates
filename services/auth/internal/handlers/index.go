package auth_handlers

import (
	"net/http"
	"html/template"
)


func Index(w http.ResponseWriter, r *http.Request, tmpl *template.Template) {
	_ = tmpl.Execute(w, map[string]any{"ShowVerify": false})
}
