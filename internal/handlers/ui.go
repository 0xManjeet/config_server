package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type UIHandler struct {
	storage  Storage
	template *template.Template
}

func NewUIHandler(storage Storage, templateDir string) (*UIHandler, error) {
	tmpl, err := template.ParseFiles(filepath.Join(templateDir, "ui.html"))
	if err != nil {
		return nil, err
	}

	return &UIHandler{
		storage:  storage,
		template: tmpl,
	}, nil
}

func (h *UIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	pass := r.URL.Query().Get("pass")

	var value string
	if key != "" {
		if data, err := h.storage.Get(key); err == nil && data != nil {
			value = string(data)
		}
	}

	data := struct {
		Key      string
		Value    string
		Password string
	}{
		Key:      key,
		Value:    value,
		Password: pass,
	}

	h.template.Execute(w, data)
} 