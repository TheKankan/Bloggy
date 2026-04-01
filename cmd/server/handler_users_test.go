package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestConfig(store *mockStore) *apiConfig {
	return &apiConfig{
		db:        store,
		jwtSecret: "test-secret",
	}
}

func makeRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ── handlerRegister ───────────────────────────────────────────────────────────

func TestHandlerRegister_Success(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeRequest(t, http.MethodPost, "/api/register", map[string]string{
		"username": "alice",
		"password": "secret123",
	})

	cfg.handlerRegister(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var resp struct {
		User  User   `json:"user"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.User.Username != "alice" {
		t.Errorf("expected username 'alice', got '%s'", resp.User.Username)
	}
	if resp.Token == "" {
		t.Error("expected a non-empty token")
	}
}

func TestHandlerRegister_MissingUsername(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeRequest(t, http.MethodPost, "/api/register", map[string]string{
		"username": "",
		"password": "secret123",
	})

	cfg.handlerRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerRegister_MissingPassword(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeRequest(t, http.MethodPost, "/api/register", map[string]string{
		"username": "alice",
		"password": "",
	})

	cfg.handlerRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerRegister_InvalidJSON(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")

	cfg.handlerRegister(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

// ── handlerLogin ──────────────────────────────────────────────────────────────

func TestHandlerLogin_Success(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	cfg.handlerRegister(httptest.NewRecorder(), makeRequest(t, http.MethodPost, "/api/register", map[string]string{
		"username": "alice",
		"password": "secret123",
	}))

	w := httptest.NewRecorder()
	cfg.handlerLogin(w, makeRequest(t, http.MethodPost, "/api/login", map[string]string{
		"username": "alice",
		"password": "secret123",
	}))

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		User  User   `json:"user"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.User.Username != "alice" {
		t.Errorf("expected username 'alice', got '%s'", resp.User.Username)
	}
	if resp.Token == "" {
		t.Error("expected a non-empty token")
	}
}

func TestHandlerLogin_WrongPassword(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	cfg.handlerRegister(httptest.NewRecorder(), makeRequest(t, http.MethodPost, "/api/register", map[string]string{
		"username": "alice",
		"password": "secret123",
	}))

	w := httptest.NewRecorder()
	cfg.handlerLogin(w, makeRequest(t, http.MethodPost, "/api/login", map[string]string{
		"username": "alice",
		"password": "wrongpassword",
	}))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerLogin_UserNotFound(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	cfg.handlerLogin(w, makeRequest(t, http.MethodPost, "/api/login", map[string]string{
		"username": "nobody",
		"password": "secret123",
	}))

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerLogin_MissingCredentials(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	cfg.handlerLogin(w, makeRequest(t, http.MethodPost, "/api/login", map[string]string{
		"username": "",
		"password": "",
	}))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
