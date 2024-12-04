package main

import (
	"bytes"
	"encoding/json"
	"jsonserver/internal/handlers"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testStorage struct {
	data map[string]json.RawMessage
}

func newTestStorage() *testStorage {
	return &testStorage{
		data: make(map[string]json.RawMessage),
	}
}

func (s *testStorage) Get(key string) (json.RawMessage, error) {
	if value, exists := s.data[key]; exists {
		return value, nil
	}
	return nil, nil
}

func (s *testStorage) Set(key string, value json.RawMessage) error {
	s.data[key] = value
	return nil
}

func TestHandleGet(t *testing.T) {
	store := newTestStorage()
	handler := handlers.NewAPIHandler(store, "test-password")

	tests := []struct {
		name         string
		key          string
		setupData    interface{}
		expectedCode int
		expectedBody string
		expectJSON   bool
	}{
		{
			name:         "get_nonexistent_key",
			key:          "nonexistent",
			expectedCode: http.StatusNotFound,
			expectedBody: "Key not found\n",
			expectJSON:   false,
		},
		{
			name:         "get_existing_key",
			key:          "test-key",
			setupData:    map[string]string{"message": "hello"},
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"hello"}`,
			expectJSON:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test data if needed
			if tt.setupData != nil {
				data, _ := json.Marshal(tt.setupData)
				store.Set(tt.key, data)
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/"+tt.key, nil)
			w := httptest.NewRecorder()

			// Handle request
			handler.HandleRequest(w, req)

			// Check status code
			if w.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			// Check response body
			if tt.expectJSON {
				if w.Header().Get("Content-Type") != "application/json" {
					t.Error("expected Content-Type to be application/json")
				}
			}
			if got := w.Body.String(); got != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, got)
			}
		})
	}
}

func TestHandlePost(t *testing.T) {
	store := newTestStorage()
	handler := handlers.NewAPIHandler(store, "test-password")

	tests := []struct {
		name         string
		key          string
		body         interface{}
		password     string
		expectedCode int
	}{
		{
			name:         "post_without_password",
			key:          "test-key",
			body:         map[string]string{"message": "hello"},
			password:     "",
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "post_wrong_password",
			key:          "test-key",
			body:         map[string]string{"message": "hello"},
			password:     "wrong-password",
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "post_valid",
			key:          "test-key",
			body:         map[string]string{"message": "hello"},
			password:     "test-password",
			expectedCode: http.StatusOK,
		},
		{
			name:         "post_invalid_json",
			key:          "test-key",
			body:         "invalid json",
			password:     "test-password",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("failed to marshal body: %v", err)
				}
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/"+tt.key, bytes.NewReader(body))
			if tt.password != "" {
				req.Header.Set("Authorization", tt.password)
			}
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Handle request
			handler.HandleRequest(w, req)

			// Check status code
			if w.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, w.Code)
			}

			// If successful POST, verify data was stored
			if tt.expectedCode == http.StatusOK {
				stored, err := store.Get(tt.key)
				if err != nil {
					t.Errorf("failed to get stored data: %v", err)
				}
				if stored == nil {
					t.Error("data was not stored")
				}
				var storedData map[string]interface{}
				if err := json.Unmarshal(stored, &storedData); err != nil {
					t.Errorf("stored data is not valid JSON: %v", err)
				}
			}
		})
	}
}

func TestCORS(t *testing.T) {
	store := newTestStorage()
	handler := handlers.NewAPIHandler(store, "test-password")

	tests := []struct {
		name          string
		origin        string
		expectHeaders bool
	}{
		{
			name:          "allowed_origin",
			origin:        "https://example.noxchat.in",
			expectHeaders: true,
		},
		{
			name:          "disallowed_origin",
			origin:        "https://example.com",
			expectHeaders: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			handler.HandleRequest(w, req)

			// Check CORS headers
			if tt.expectHeaders {
				if w.Header().Get("Access-Control-Allow-Origin") != tt.origin {
					t.Error("expected Access-Control-Allow-Origin header to be set")
				}
				if w.Header().Get("Access-Control-Allow-Methods") == "" {
					t.Error("expected Access-Control-Allow-Methods header to be set")
				}
				if w.Header().Get("Access-Control-Allow-Headers") == "" {
					t.Error("expected Access-Control-Allow-Headers header to be set")
				}
			} else {
				if w.Header().Get("Access-Control-Allow-Origin") != "" {
					t.Error("expected no Access-Control-Allow-Origin header")
				}
			}
		})
	}
}
