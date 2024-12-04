package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

type APIHandler struct {
	storage  Storage
	password string
}

type Storage interface {
	Get(key string) (json.RawMessage, error)
	Set(key string, value json.RawMessage) error
}

func NewAPIHandler(storage Storage, password string) *APIHandler {
	return &APIHandler{
		storage:  storage,
		password: password,
	}
}

func (h *APIHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	origin := r.Header.Get("Origin")
	if strings.HasSuffix(origin, "noxchat.in") {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/")
	if key == "" {
		http.Error(w, "Key required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, key)
	case http.MethodPost:
		h.handlePost(w, r, key)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) handleGet(w http.ResponseWriter, key string) {
	data, err := h.storage.Get(key)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if data == nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *APIHandler) handlePost(w http.ResponseWriter, r *http.Request, key string) {
	// Validate key format (example: alphanumeric + limited special chars only)
	if !isValidKey(key) {
		http.Error(w, "Invalid key format", http.StatusBadRequest)
		return
	}

	if r.Header.Get("Authorization") != h.password {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Set a reasonable size limit (e.g., 1MB)
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	var jsonData json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
		if err.Error() == "http: request body too large" {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.storage.Set(key, jsonData); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Add this helper function
func isValidKey(key string) bool {
	if len(key) > 256 { // Set a reasonable maximum key length
		return false
	}
	// Only allow alphanumeric characters and certain special characters
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
