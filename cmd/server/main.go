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

	// Auth
	r.Post("/register", apiCfg.handlerRegister)
	r.Post("/login", apiCfg.handlerLogin)

	// Articles
	r.Get("/articles", apiCfg.handlerGetArticles)
	r.Get("/articles/{id}", apiCfg.handlerGetArticle)
	r.Post("/articles", apiCfg.middlewareAuth(apiCfg.handlerCreateArticle))
	r.Put("/articles/{id}", apiCfg.middlewareAuth(apiCfg.handlerUpdateArticle))
	r.Delete("/articles/{id}", apiCfg.middlewareAuth(apiCfg.handlerDeleteArticle))

	// Comments
	r.Get("/articles/{id}/comments", apiCfg.handlerGetComments)
	r.Post("/articles/{id}/comments", apiCfg.middlewareAuth(apiCfg.handlerCreateComment))
	r.Delete("/articles/{id}/comments/{commentId}", apiCfg.middlewareAuth(apiCfg.handlerDeleteComment))

	log.Println("Serving on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
