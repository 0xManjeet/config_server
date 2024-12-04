package handlers

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/valyala/fasthttp"
)

type APIHandler struct {
	storage            Storage
	password           string
	corsAllowedOrigins string
	pool               sync.Pool
}

type Storage interface {
	Get(key string) (json.RawMessage, error)
	Set(key string, value json.RawMessage) error
}

func NewAPIHandler(storage Storage, password string, corsAllowedOrigins string) *APIHandler {
	return &APIHandler{
		storage:            storage,
		password:           password,
		corsAllowedOrigins: corsAllowedOrigins,
		pool: sync.Pool{
			New: func() interface{} {
				b := make([]byte, 1024*1024)
				return &b
			},
		},
	}
}

func (h *APIHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	origin := string(ctx.Request.Header.Peek("Origin"))
	if strings.HasSuffix(origin, h.corsAllowedOrigins) {
		ctx.Response.Header.Set("Access-Control-Allow-Origin", origin)
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	}

	if string(ctx.Method()) == "OPTIONS" {
		ctx.SetStatusCode(fasthttp.StatusOK)
		return
	}

	key := string(ctx.Path()[1:])
	if key == "" {
		ctx.Error("Key required", fasthttp.StatusBadRequest)
		return
	}

	switch string(ctx.Method()) {
	case "GET":
		h.handleFastGet(ctx, key)
	case "POST":
		h.handleFastPost(ctx, key)
	default:
		ctx.Error("Method not allowed", fasthttp.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) handleFastGet(ctx *fasthttp.RequestCtx, key string) {
	data, err := h.storage.Get(key)
	if err != nil {
		ctx.Error("Internal server error", fasthttp.StatusInternalServerError)
		return
	}
	if data == nil {
		ctx.Error("Key not found", fasthttp.StatusNotFound)
		return
	}

	ctx.SetContentType("application/json")
	ctx.Write(data)
}

func (h *APIHandler) handleFastPost(ctx *fasthttp.RequestCtx, key string) {
	if !isValidKey(key) {
		ctx.Error("Invalid key format", fasthttp.StatusBadRequest)
		return
	}

	if string(ctx.Request.Header.Peek("Authorization")) != h.password {
		ctx.Error("Unauthorized", fasthttp.StatusForbidden)
		return
	}

	bufferPtr := h.pool.Get().(*[]byte)
	defer h.pool.Put(bufferPtr)
	buffer := *bufferPtr

	if len(ctx.PostBody()) > len(buffer) {
		ctx.Error("Request body too large", fasthttp.StatusRequestEntityTooLarge)
		return
	}

	if !json.Valid(ctx.PostBody()) {
		ctx.Error("Invalid JSON", fasthttp.StatusBadRequest)
		return
	}

	if err := h.storage.Set(key, json.RawMessage(ctx.PostBody())); err != nil {
		ctx.Error("Internal server error", fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func isValidKey(key string) bool {
	if len(key) > 256 {
		return false
	}
	for _, r := range key {
		if !isAllowedKeyChar(r) {
			return false
		}
	}
	return true
}

func isAllowedKeyChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_' || r == '.'
}
