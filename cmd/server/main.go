package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/TheKankan/Bloggy/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db        database.Store
	jwtSecret string
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	apiCfg := apiConfig{
		db:        database.New(dbConn),
		jwtSecret: jwtSecret,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Auth API
	r.Post("/api/register", apiCfg.handlerRegister)
	r.Post("/api/login", apiCfg.handlerLogin)

	// Articles API
	r.Get("/api/articles", apiCfg.handlerGetArticles)
	r.Get("/api/articles/{id}", apiCfg.handlerGetArticle)
	r.Post("/api/articles", apiCfg.middlewareAuth(apiCfg.handlerCreateArticle))
	r.Put("/api/articles/{id}", apiCfg.middlewareAuth(apiCfg.handlerUpdateArticle))
	r.Delete("/api/articles/{id}", apiCfg.middlewareAuth(apiCfg.handlerDeleteArticle))

	// Comments API
	r.Get("/api/articles/{id}/comments", apiCfg.handlerGetComments)
	r.Post("/api/articles/{id}/comments", apiCfg.middlewareAuth(apiCfg.handlerCreateComment))
	r.Delete("/api/articles/{id}/comments/{commentId}", apiCfg.middlewareAuth(apiCfg.handlerDeleteComment))

	// Web UI
	r.Get("/", apiCfg.webHandlerHome)
	r.Get("/login", apiCfg.webHandlerLoginPage)
	r.Post("/login", apiCfg.webHandlerLoginSubmit)
	r.Get("/register", apiCfg.webHandlerRegisterPage)
	r.Post("/register", apiCfg.webHandlerRegisterSubmit)
	r.Get("/articles/new", apiCfg.webHandlerNewArticlePage)
	r.Post("/articles/new", apiCfg.webHandlerNewArticleSubmit)
	r.Get("/web/articles/{id}", apiCfg.webHandlerArticlePage)
	r.Post("/web/articles/{id}/comments", apiCfg.webHandlerCreateComment)
	r.Post("/web/articles/{id}/delete", apiCfg.webHandlerDeleteArticle)
	r.Get("/logout", apiCfg.webHandlerLogout)
	r.Post("/web/articles/{id}/comments/{commentId}/delete", apiCfg.webHandlerDeleteComment)

	log.Println("Serving on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
