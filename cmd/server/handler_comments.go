package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/TheKankan/Bloggy/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Content   string    `json:"content"`
	UserID    uuid.UUID `json:"user_id"`
	ArticleID uuid.UUID `json:"article_id"`
}

func (cfg *apiConfig) handlerGetComments(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid article ID", err)
		return
	}

	// Check article exists
	_, err = cfg.db.GetArticleByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Article not found", err)
		return
	}

	comments, err := cfg.db.GetCommentsByArticleID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get comments", err)
		return
	}

	resp := make([]Comment, len(comments))
	for i, c := range comments {
		resp[i] = Comment{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Content:   c.Content,
			UserID:    c.UserID,
			ArticleID: c.ArticleID,
		}
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handlerCreateComment(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Content string `json:"content"`
	}

	userID, ok := getUserIDFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	articleID, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid article ID", err)
		return
	}

	// Check article exists
	_, err = cfg.db.GetArticleByID(r.Context(), articleID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Article not found", err)
		return
	}

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Content == "" {
		respondWithError(w, http.StatusBadRequest, "Content is required", nil)
		return
	}

	comment, err := cfg.db.CreateComment(r.Context(), database.CreateCommentParams{
		Content:   params.Content,
		UserID:    userID,
		ArticleID: articleID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create comment", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Comment{
		ID:        comment.ID,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
		Content:   comment.Content,
		UserID:    comment.UserID,
		ArticleID: comment.ArticleID,
	})
}

func (cfg *apiConfig) handlerDeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	commentIDStr := chi.URLParam(r, "commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid comment ID", err)
		return
	}

	// Check comment exists and belongs to user
	comment, err := cfg.db.GetCommentByID(r.Context(), commentID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Comment not found", err)
		return
	}
	if comment.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You can only delete your own comments", nil)
		return
	}

	if err := cfg.db.DeleteComment(r.Context(), commentID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete comment", err)
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{Message: "Comment deleted successfully"})
}
