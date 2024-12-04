package handlers

import (
	"html/template"
	"path/filepath"

	"github.com/valyala/fasthttp"
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

func (h *UIHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/html; charset=utf-8")

	key := string(ctx.QueryArgs().Peek("key"))
	pass := string(ctx.QueryArgs().Peek("pass"))

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

	h.template.Execute(ctx, data)
}
