package database

import (
	"context"

	"github.com/google/uuid"
)

type Store interface {
	// Users
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetUserFromUsername(ctx context.Context, username string) (User, error)

	// Articles
	CreateArticle(ctx context.Context, arg CreateArticleParams) (Article, error)
	GetArticles(ctx context.Context, arg GetArticlesParams) ([]Article, error)
	GetArticleByID(ctx context.Context, id uuid.UUID) (Article, error)
	UpdateArticle(ctx context.Context, arg UpdateArticleParams) (Article, error)
	DeleteArticle(ctx context.Context, id uuid.UUID) error

	// Comments
	CreateComment(ctx context.Context, arg CreateCommentParams) (Comment, error)
	GetCommentsByArticleID(ctx context.Context, articleID uuid.UUID) ([]Comment, error)
	DeleteComment(ctx context.Context, id uuid.UUID) error
	GetCommentByID(ctx context.Context, id uuid.UUID) (Comment, error)
}
