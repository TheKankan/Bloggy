package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TheKankan/Bloggy/internal/database"
	"github.com/google/uuid"
)

// ── handlerGetComments ────────────────────────────────────────────────────────

func TestHandlerGetComments_Empty(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)
	articleID := uuid.New()
	store.articles[articleID] = makeTestArticle("Title", "Content", userID)

	w := httptest.NewRecorder()
	req := makeChiRequest(
		httptest.NewRequest(http.MethodGet, "/api/articles/"+articleID.String()+"/comments", nil),
		map[string]string{"id": articleID.String()},
	)
	cfg.handlerGetComments(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp []Comment
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d comments", len(resp))
	}
}

func TestHandlerGetComments_ArticleNotFound(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeChiRequest(
		httptest.NewRequest(http.MethodGet, "/api/articles/"+uuid.New().String()+"/comments", nil),
		map[string]string{"id": uuid.New().String()},
	)
	cfg.handlerGetComments(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestHandlerGetComments_InvalidID(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeChiRequest(
		httptest.NewRequest(http.MethodGet, "/api/articles/not-a-uuid/comments", nil),
		map[string]string{"id": "not-a-uuid"},
	)
	cfg.handlerGetComments(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

// ── handlerCreateComment ──────────────────────────────────────────────────────

func TestHandlerCreateComment_Success(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)
	articleID := uuid.New()
	store.articles[articleID] = makeTestArticle("Title", "Content", userID)

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodPost, "/api/articles/"+articleID.String()+"/comments", map[string]string{
		"content": "Great article!",
	}, userID, cfg.jwtSecret)
	req = makeChiRequest(req, map[string]string{"id": articleID.String()})

	cfg.middlewareAuth(cfg.handlerCreateComment)(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var resp Comment
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Content != "Great article!" {
		t.Errorf("expected content 'Great article!', got '%s'", resp.Content)
	}
}

func TestHandlerCreateComment_MissingContent(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)
	articleID := uuid.New()
	store.articles[articleID] = makeTestArticle("Title", "Content", userID)

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodPost, "/api/articles/"+articleID.String()+"/comments", map[string]string{
		"content": "",
	}, userID, cfg.jwtSecret)
	req = makeChiRequest(req, map[string]string{"id": articleID.String()})

	cfg.middlewareAuth(cfg.handlerCreateComment)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerCreateComment_ArticleNotFound(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodPost, "/api/articles/"+uuid.New().String()+"/comments", map[string]string{
		"content": "Great article!",
	}, userID, cfg.jwtSecret)
	req = makeChiRequest(req, map[string]string{"id": uuid.New().String()})

	cfg.middlewareAuth(cfg.handlerCreateComment)(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestHandlerCreateComment_Unauthorized(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeRequest(t, http.MethodPost, "/api/articles/"+uuid.New().String()+"/comments", map[string]string{
		"content": "Great article!",
	})

	cfg.middlewareAuth(cfg.handlerCreateComment)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

// ── handlerDeleteComment ──────────────────────────────────────────────────────

func TestHandlerDeleteComment_Success(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)
	articleID := uuid.New()
	store.articles[articleID] = makeTestArticle("Title", "Content", userID)
	commentID := uuid.New()
	store.comments[commentID] = database.Comment{
		ID:        commentID,
		Content:   "Great article!",
		UserID:    userID,
		ArticleID: articleID,
	}

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodDelete, "/api/articles/"+articleID.String()+"/comments/"+commentID.String(), nil, userID, cfg.jwtSecret)
	req = makeChiRequest(req, map[string]string{
		"id":        articleID.String(),
		"commentId": commentID.String(),
	})

	cfg.middlewareAuth(cfg.handlerDeleteComment)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if _, exists := store.comments[commentID]; exists {
		t.Error("expected comment to be deleted")
	}
}

func TestHandlerDeleteComment_Forbidden(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	ownerID := uuid.New()
	otherID := uuid.New()
	store.users["alice"] = makeTestUser("alice", ownerID)
	store.users["bob"] = makeTestUser("bob", otherID)
	articleID := uuid.New()
	store.articles[articleID] = makeTestArticle("Title", "Content", ownerID)
	commentID := uuid.New()
	store.comments[commentID] = database.Comment{
		ID:        commentID,
		Content:   "Great article!",
		UserID:    ownerID,
		ArticleID: articleID,
	}

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodDelete, "/api/articles/"+articleID.String()+"/comments/"+commentID.String(), nil, otherID, cfg.jwtSecret)
	req = makeChiRequest(req, map[string]string{
		"id":        articleID.String(),
		"commentId": commentID.String(),
	})

	cfg.middlewareAuth(cfg.handlerDeleteComment)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestHandlerDeleteComment_NotFound(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	userID := uuid.New()
	store.users["alice"] = makeTestUser("alice", userID)

	w := httptest.NewRecorder()
	req := makeAuthRequest(t, http.MethodDelete, "/api/articles/"+uuid.New().String()+"/comments/"+uuid.New().String(), nil, userID, cfg.jwtSecret)
	req = makeChiRequest(req, map[string]string{
		"id":        uuid.New().String(),
		"commentId": uuid.New().String(),
	})

	cfg.middlewareAuth(cfg.handlerDeleteComment)(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}
