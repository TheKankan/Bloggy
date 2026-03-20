package main

import (
	"html/template"
	"net/http"
	"time"

	"github.com/TheKankan/Bloggy/internal/auth"
	"github.com/TheKankan/Bloggy/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func renderTemplate(w http.ResponseWriter, page string, data interface{}) {
	ts, err := template.ParseFiles("web/templates/base.html", "web/templates/"+page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ts.ExecuteTemplate(w, "base.html", data)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func getUserIDFromCookie(r *http.Request, jwtSecret string) (uuid.UUID, bool) {
	cookie, err := r.Cookie("token")
	if err != nil {
		return uuid.Nil, false
	}
	userID, err := auth.ValidateJWT(cookie.Value, jwtSecret)
	if err != nil {
		return uuid.Nil, false
	}
	return userID, true
}

func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(24 * time.Hour / time.Second),
	})
}

func clearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
}

// ── Auth pages ────────────────────────────────────────────────────────────────

func (cfg *apiConfig) webHandlerLoginPage(w http.ResponseWriter, r *http.Request) {
	_, loggedIn := getUserIDFromCookie(r, cfg.jwtSecret)
	renderTemplate(w, "login.html", map[string]interface{}{
		"LoggedIn": loggedIn,
	})
}

func (cfg *apiConfig) webHandlerLoginSubmit(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := cfg.db.GetUserFromUsername(r.Context(), username)
	if err != nil {
		renderTemplate(w, "login.html", map[string]interface{}{
			"Error":    "Invalid credentials",
			"LoggedIn": false,
		})
		return
	}

	match, err := auth.CheckPasswordHash(password, user.HashedPassword)
	if err != nil || !match {
		renderTemplate(w, "login.html", map[string]interface{}{
			"Error":    "Invalid credentials",
			"LoggedIn": false,
		})
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, 24*time.Hour)
	if err != nil {
		renderTemplate(w, "login.html", map[string]interface{}{
			"Error":    "Could not create session",
			"LoggedIn": false,
		})
		return
	}

	setTokenCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (cfg *apiConfig) webHandlerRegisterPage(w http.ResponseWriter, r *http.Request) {
	_, loggedIn := getUserIDFromCookie(r, cfg.jwtSecret)
	renderTemplate(w, "register.html", map[string]interface{}{
		"LoggedIn": loggedIn,
	})
}

func (cfg *apiConfig) webHandlerRegisterSubmit(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		renderTemplate(w, "register.html", map[string]interface{}{
			"Error":    "Could not hash password",
			"LoggedIn": false,
		})
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Username:       username,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		renderTemplate(w, "register.html", map[string]interface{}{
			"Error":    "Username already taken",
			"LoggedIn": false,
		})
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, 24*time.Hour)
	if err != nil {
		renderTemplate(w, "register.html", map[string]interface{}{
			"Error":    "Could not create session",
			"LoggedIn": false,
		})
		return
	}

	setTokenCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (cfg *apiConfig) webHandlerLogout(w http.ResponseWriter, r *http.Request) {
	clearTokenCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ── Home ──────────────────────────────────────────────────────────────────────

type ArticleWithUsername struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Title     string
	Username  string
}

func (cfg *apiConfig) webHandlerHome(w http.ResponseWriter, r *http.Request) {
	articles, err := cfg.db.GetArticles(r.Context(), database.GetArticlesParams{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		http.Error(w, "Could not get articles", http.StatusInternalServerError)
		return
	}

	type pageData struct {
		Articles []ArticleWithUsername
		LoggedIn bool
	}

	_, loggedIn := getUserIDFromCookie(r, cfg.jwtSecret)

	data := pageData{LoggedIn: loggedIn}
	for _, a := range articles {
		username, _ := cfg.db.GetUsernameFromID(r.Context(), a.UserID)
		data.Articles = append(data.Articles, ArticleWithUsername{
			ID:        a.ID,
			CreatedAt: a.CreatedAt,
			Title:     a.Title,
			Username:  username,
		})
	}

	renderTemplate(w, "home.html", data)
}

// ── Article page ──────────────────────────────────────────────────────────────

type ArticleDetail struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Title     string
	Content   string
	Username  string
	UserID    uuid.UUID
}

type CommentDetail struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Content   string
	Username  string
	UserID    uuid.UUID
}

func (cfg *apiConfig) webHandlerArticlePage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	article, err := cfg.db.GetArticleByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	username, _ := cfg.db.GetUsernameFromID(r.Context(), article.UserID)

	comments, err := cfg.db.GetCommentsByArticleID(r.Context(), id)
	if err != nil {
		http.Error(w, "Could not get comments", http.StatusInternalServerError)
		return
	}

	userID, loggedIn := getUserIDFromCookie(r, cfg.jwtSecret)

	var commentDetails []CommentDetail
	for _, c := range comments {
		commentUsername, _ := cfg.db.GetUsernameFromID(r.Context(), c.UserID)
		commentDetails = append(commentDetails, CommentDetail{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			Content:   c.Content,
			Username:  commentUsername,
			UserID:    c.UserID,
		})
	}

	renderTemplate(w, "article.html", map[string]interface{}{
		"Article": ArticleDetail{
			ID:        article.ID,
			CreatedAt: article.CreatedAt,
			Title:     article.Title,
			Content:   article.Content,
			Username:  username,
			UserID:    article.UserID,
		},
		"Comments":      commentDetails,
		"LoggedIn":      loggedIn,
		"IsAuthor":      loggedIn && userID == article.UserID,
		"CurrentUserID": userID.String(),
	})
}

// ── New article ───────────────────────────────────────────────────────────────

func (cfg *apiConfig) webHandlerNewArticlePage(w http.ResponseWriter, r *http.Request) {
	_, loggedIn := getUserIDFromCookie(r, cfg.jwtSecret)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	renderTemplate(w, "new_article.html", map[string]interface{}{
		"LoggedIn": true,
	})
}

func (cfg *apiConfig) webHandlerNewArticleSubmit(w http.ResponseWriter, r *http.Request) {
	userID, loggedIn := getUserIDFromCookie(r, cfg.jwtSecret)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		renderTemplate(w, "new_article.html", map[string]interface{}{
			"Error":    "Title and content are required",
			"LoggedIn": true,
		})
		return
	}

	article, err := cfg.db.CreateArticle(r.Context(), database.CreateArticleParams{
		Title:   title,
		Content: content,
		UserID:  userID,
	})
	if err != nil {
		renderTemplate(w, "new_article.html", map[string]interface{}{
			"Error":    "Could not create article",
			"LoggedIn": true,
		})
		return
	}

	http.Redirect(w, r, "/web/articles/"+article.ID.String(), http.StatusSeeOther)
}

// ── Comment submit ────────────────────────────────────────────────────────────

func (cfg *apiConfig) webHandlerCreateComment(w http.ResponseWriter, r *http.Request) {
	userID, loggedIn := getUserIDFromCookie(r, cfg.jwtSecret)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	idStr := chi.URLParam(r, "id")
	articleID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Redirect(w, r, "/web/articles/"+idStr, http.StatusSeeOther)
		return
	}

	cfg.db.CreateComment(r.Context(), database.CreateCommentParams{
		Content:   content,
		UserID:    userID,
		ArticleID: articleID,
	})

	http.Redirect(w, r, "/web/articles/"+idStr, http.StatusSeeOther)
}

// ── Delete article ────────────────────────────────────────────────────────────

func (cfg *apiConfig) webHandlerDeleteArticle(w http.ResponseWriter, r *http.Request) {
	userID, loggedIn := getUserIDFromCookie(r, cfg.jwtSecret)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	article, err := cfg.db.GetArticleByID(r.Context(), id)
	if err != nil || article.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	cfg.db.DeleteArticle(r.Context(), id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ── Delete comment ────────────────────────────────────────────────────────────

func (cfg *apiConfig) webHandlerDeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, loggedIn := getUserIDFromCookie(r, cfg.jwtSecret)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	commentIDStr := chi.URLParam(r, "commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	comment, err := cfg.db.GetCommentByID(r.Context(), commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}
	article, err := cfg.db.GetArticleByID(r.Context(), comment.ArticleID)
	if err != nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	isCommentAuthor := comment.UserID == userID
	isArticleAuthor := article.UserID == userID

	if !isCommentAuthor && !isArticleAuthor {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	cfg.db.DeleteComment(r.Context(), commentID)

	articleIDStr := chi.URLParam(r, "id")
	http.Redirect(w, r, "/web/articles/"+articleIDStr, http.StatusSeeOther)
}
