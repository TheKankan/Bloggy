package main

import (
	"context"
	"errors"
	"time"

	"github.com/TheKankan/Bloggy/internal/database"
	"github.com/google/uuid"
)

type mockStore struct {
	users    map[string]database.User
	articles map[uuid.UUID]database.Article
	comments map[uuid.UUID]database.Comment
}

func newMockStore() *mockStore {
	return &mockStore{
		users:    make(map[string]database.User),
		articles: make(map[uuid.UUID]database.Article),
		comments: make(map[uuid.UUID]database.Comment),
	}
}

// ── Users ─────────────────────────────────────────────────────────────────────

func (m *mockStore) CreateUser(_ context.Context, arg database.CreateUserParams) (database.User, error) {
	if _, exists := m.users[arg.Username]; exists {
		return database.User{}, errors.New("username already taken")
	}
	user := database.User{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Username:       arg.Username,
		HashedPassword: arg.HashedPassword,
	}
	m.users[arg.Username] = user
	return user, nil
}

func (m *mockStore) GetUserFromUsername(_ context.Context, username string) (database.User, error) {
	user, ok := m.users[username]
	if !ok {
		return database.User{}, errors.New("user not found")
	}
	return user, nil
}

func (m *mockStore) GetUsernameFromID(_ context.Context, id uuid.UUID) (string, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u.Username, nil
		}
	}
	return "", errors.New("user not found")
}

// ── Articles ──────────────────────────────────────────────────────────────────

func (m *mockStore) CreateArticle(_ context.Context, arg database.CreateArticleParams) (database.Article, error) {
	article := database.Article{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Title:     arg.Title,
		Content:   arg.Content,
		UserID:    arg.UserID,
	}
	m.articles[article.ID] = article
	return article, nil
}

func (m *mockStore) GetArticles(_ context.Context, arg database.GetArticlesParams) ([]database.Article, error) {
	articles := make([]database.Article, 0, len(m.articles))
	for _, a := range m.articles {
		articles = append(articles, a)
	}

	// Apply offset and limit
	offset := int(arg.Offset)
	if offset > len(articles) {
		return []database.Article{}, nil
	}
	articles = articles[offset:]

	limit := int(arg.Limit)
	if limit < len(articles) {
		articles = articles[:limit]
	}

	return articles, nil
}

func (m *mockStore) GetArticleByID(_ context.Context, id uuid.UUID) (database.Article, error) {
	article, ok := m.articles[id]
	if !ok {
		return database.Article{}, errors.New("article not found")
	}
	return article, nil
}

func (m *mockStore) UpdateArticle(_ context.Context, arg database.UpdateArticleParams) (database.Article, error) {
	article, ok := m.articles[arg.ID]
	if !ok {
		return database.Article{}, errors.New("article not found")
	}
	article.Title = arg.Title
	article.Content = arg.Content
	article.UpdatedAt = time.Now()
	m.articles[arg.ID] = article
	return article, nil
}

func (m *mockStore) DeleteArticle(_ context.Context, id uuid.UUID) error {
	if _, ok := m.articles[id]; !ok {
		return errors.New("article not found")
	}
	delete(m.articles, id)
	return nil
}

// ── Comments ──────────────────────────────────────────────────────────────────

func (m *mockStore) CreateComment(_ context.Context, arg database.CreateCommentParams) (database.Comment, error) {
	comment := database.Comment{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Content:   arg.Content,
		UserID:    arg.UserID,
		ArticleID: arg.ArticleID,
	}
	m.comments[comment.ID] = comment
	return comment, nil
}

func (m *mockStore) GetCommentsByArticleID(_ context.Context, articleID uuid.UUID) ([]database.Comment, error) {
	comments := []database.Comment{}
	for _, c := range m.comments {
		if c.ArticleID == articleID {
			comments = append(comments, c)
		}
	}
	return comments, nil
}

func (m *mockStore) GetCommentByID(_ context.Context, id uuid.UUID) (database.Comment, error) {
	comment, ok := m.comments[id]
	if !ok {
		return database.Comment{}, errors.New("comment not found")
	}
	return comment, nil
}

func (m *mockStore) DeleteComment(_ context.Context, id uuid.UUID) error {
	if _, ok := m.comments[id]; !ok {
		return errors.New("comment not found")
	}
	delete(m.comments, id)
	return nil
}
