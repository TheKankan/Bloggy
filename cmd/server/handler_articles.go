package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/TheKankan/Bloggy/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Article struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerGetArticles(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := int32(10) // default
	offset := int32(0) // default

	if limitStr != "" {
		val, err := strconv.Atoi(limitStr)
		if err != nil || val <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid limit parameter", nil)
			return
		}
		limit = int32(val)
	}

	if offsetStr != "" {
		val, err := strconv.Atoi(offsetStr)
		if err != nil || val < 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid offset parameter", nil)
			return
		}
		offset = int32(val)
	}

	articles, err := cfg.db.GetArticles(r.Context(), database.GetArticlesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get articles", err)
		return
	}

	resp := make([]Article, len(articles))
	for i, a := range articles {
		resp[i] = Article{
			ID:        a.ID,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			Title:     a.Title,
			Content:   a.Content,
			UserID:    a.UserID,
		}
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handlerGetArticle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid article ID", err)
		return
	}

	article, err := cfg.db.GetArticleByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Article not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Article{
		ID:        article.ID,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
		Title:     article.Title,
		Content:   article.Content,
		UserID:    article.UserID,
	})
}

func (cfg *apiConfig) handlerCreateArticle(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	userID, ok := getUserIDFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Title == "" || params.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Title and content are required", nil)
		return
	}

	article, err := cfg.db.CreateArticle(r.Context(), database.CreateArticleParams{
		Title:   params.Title,
		Content: params.Content,
		UserID:  userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create article", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Article{
		ID:        article.ID,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
		Title:     article.Title,
		Content:   article.Content,
		UserID:    article.UserID,
	})
}

func (cfg *apiConfig) handlerUpdateArticle(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	userID, ok := getUserIDFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid article ID", err)
		return
	}

	// Check article exists and belongs to user
	article, err := cfg.db.GetArticleByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Article not found", err)
		return
	}
	if article.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You can only edit your own articles", nil)
		return
	}

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Title == "" || params.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Title and content are required", nil)
		return
	}

	updated, err := cfg.db.UpdateArticle(r.Context(), database.UpdateArticleParams{
		ID:      id,
		Title:   params.Title,
		Content: params.Content,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update article", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Article{
		ID:        updated.ID,
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
		Title:     updated.Title,
		Content:   updated.Content,
		UserID:    updated.UserID,
	})
}

func (cfg *apiConfig) handlerDeleteArticle(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid article ID", err)
		return
	}

	// Check article exists and belongs to user
	article, err := cfg.db.GetArticleByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Article not found", err)
		return
	}
	if article.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You can only delete your own articles", nil)
		return
	}

	if err := cfg.db.DeleteArticle(r.Context(), id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete article", err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: "Article deleted successfully"})
}
