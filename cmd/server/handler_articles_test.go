package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TheKankan/Bloggy/internal/auth"
	"github.com/TheKankan/Bloggy/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// makeAuthRequest crée une requête avec un token JWT valide
func makeAuthRequest(t *testing.T, method, path string, body interface{}, userID uuid.UUID, secret string) *http.Request {
	t.Helper()
	req := makeRequest(t, method, path, body)
	token, err := auth.MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("failed to make JWT: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

// makeChiRequest injecte les paramètres d'URL chi dans la requête
func makeChiRequest(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// ── handlerGetArticles ────────────────────────────────────────────────────────

func TestHandlerGetArticles_Empty(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
	cfg.handlerGetArticles(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []Article
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d articles", len(resp))
	}
}

func TestHandlerGetArticles_WithArticles(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	// Register a user and create an article
	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)
	store.articles[uuid.New()] = makeTestArticle("Title 1", "Content 1", userID)
	store.articles[uuid.New()] = makeTestArticle("Title 2", "Content 2", userID)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
	cfg.handlerGetArticles(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []Article
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 articles, got %d", len(resp))
	}
}

func TestHandlerGetArticles_InvalidLimit(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/api/articles?limit=abc", nil)
	cfg.handlerGetArticles(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

// ── handlerGetArticle ─────────────────────────────────────────────────────────

func TestHandlerGetArticle_Success(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)
	articleID := uuid.New()
	store.articles[articleID] = makeTestArticle("My Title", "My Content", userID)

	w := httptest.NewRecorder()
	req := makeChiRequest(
		httptest.NewRequest(http.MethodGet, "/api/articles/"+articleID.String(), nil),
		map[string]string{"id": articleID.String()},
	)
	cfg.handlerGetArticle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp Article
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Title != "My Title" {
		t.Errorf("expected title 'My Title', got '%s'", resp.Title)
	}
}

func TestHandlerGetArticle_NotFound(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeChiRequest(
		httptest.NewRequest(http.MethodGet, "/api/articles/"+uuid.New().String(), nil),
		map[string]string{"id": uuid.New().String()},
	)
	cfg.handlerGetArticle(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestHandlerGetArticle_InvalidID(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeChiRequest(
		httptest.NewRequest(http.MethodGet, "/api/articles/not-a-uuid", nil),
		map[string]string{"id": "not-a-uuid"},
	)
	cfg.handlerGetArticle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

// ── handlerCreateArticle ──────────────────────────────────────────────────────

func TestHandlerCreateArticle_Success(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodPost, "/api/articles", map[string]string{
		"title":   "My Article",
		"content": "Some content here",
	}, userID, cfg.jwtSecret)

	cfg.middlewareAuth(cfg.handlerCreateArticle)(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var resp Article
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Title != "My Article" {
		t.Errorf("expected title 'My Article', got '%s'", resp.Title)
	}
}

func TestHandlerCreateArticle_MissingTitle(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodPost, "/api/articles", map[string]string{
		"title":   "",
		"content": "Some content",
	}, userID, cfg.jwtSecret)

	cfg.middlewareAuth(cfg.handlerCreateArticle)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerCreateArticle_Unauthorized(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeRequest(t, http.MethodPost, "/api/articles", map[string]string{
		"title":   "My Article",
		"content": "Some content",
	})

	cfg.middlewareAuth(cfg.handlerCreateArticle)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

// ── handlerUpdateArticle ──────────────────────────────────────────────────────

func TestHandlerUpdateArticle_Success(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)
	articleID := uuid.New()
	store.articles[articleID] = makeTestArticle("Old Title", "Old Content", userID)

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodPut, "/api/articles/"+articleID.String(), map[string]string{
		"title":   "New Title",
		"content": "New Content",
	}, userID, cfg.jwtSecret)
	req = makeChiRequest(req, map[string]string{"id": articleID.String()})

	cfg.middlewareAuth(cfg.handlerUpdateArticle)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp Article
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", resp.Title)
	}
}

func TestHandlerUpdateArticle_Forbidden(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	ownerID := uuid.New()
	otherID := uuid.New()
	store.users["alice"] = makeTestUser("alice", ownerID)
	store.users["bob"] = makeTestUser("bob", otherID)
	articleID := uuid.New()
	store.articles[articleID] = makeTestArticle("Title", "Content", ownerID)

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodPut, "/api/articles/"+articleID.String(), map[string]string{
		"title":   "Hacked Title",
		"content": "Hacked Content",
	}, otherID, cfg.jwtSecret)
	req = makeChiRequest(req, map[string]string{"id": articleID.String()})

	cfg.middlewareAuth(cfg.handlerUpdateArticle)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

// ── handlerDeleteArticle ──────────────────────────────────────────────────────

func TestHandlerDeleteArticle_Success(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)
	articleID := uuid.New()
	store.articles[articleID] = makeTestArticle("Title", "Content", userID)

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodDelete, "/api/articles/"+articleID.String(), nil, userID, cfg.jwtSecret)
	req = makeChiRequest(req, map[string]string{"id": articleID.String()})

	cfg.middlewareAuth(cfg.handlerDeleteArticle)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if _, exists := store.articles[articleID]; exists {
		t.Error("expected article to be deleted")
	}
}

func TestHandlerDeleteArticle_Forbidden(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	ownerID := uuid.New()
	otherID := uuid.New()
	store.users["alice"] = makeTestUser("alice", ownerID)
	store.users["bob"] = makeTestUser("bob", otherID)
	articleID := uuid.New()
	store.articles[articleID] = makeTestArticle("Title", "Content", ownerID)

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodDelete, "/api/articles/"+articleID.String(), nil, otherID, cfg.jwtSecret)
	req = makeChiRequest(req, map[string]string{"id": articleID.String()})

	cfg.middlewareAuth(cfg.handlerDeleteArticle)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func makeTestUser(username string, id uuid.UUID) database.User {
	return database.User{
		ID:             id,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Username:       username,
		HashedPassword: "hashedpassword",
	}
}

func makeTestArticle(title, content string, userID uuid.UUID) database.Article {
	return database.Article{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Title:     title,
		Content:   content,
		UserID:    userID,
	}
}
