![CI](https://github.com/TheKankan/Bloggy/actions/workflows/ci.yml/badge.svg)

# Bloggy

A REST API and web interface for a blog platform, built in Go. Users can register, write articles and comment on each other's posts.

## Features

- 🔐 Authentication with hashed passwords (argon2id) and JWT tokens
- 📝 Full CRUD for articles and comments
- 🛡️ Route protection with auth middleware
- 🌐 Web interface served with Go templates

## Quick Start (Docker)

### Prerequisites

- [Docker](https://www.docker.com/get-started) installed

### Running

1. Clone the repo
```bash
git clone https://github.com/TheKankan/Terminal_Chat.git
```

2. Edit .env.example : rename it .env and change the values inside it

3. Start the server and database
```bash
docker compose up
```

The server is now running on `http://localhost:8080` !

## Manual Setup (without Docker)

### Prerequisites

- Go 1.25+
- PostgreSQL
- [goose](https://github.com/pressly/goose) for migrations

### Running

1. Clone the repo
```bash
git clone https://github.com/TheKankan/Bloggy.git
```
2. Edit .env with your PostgreSQL connection string and JWT_SECRET

3. Run migrations
```bash
goose -dir sql/schema postgres "$DB_URL" up
```

4. Start the server
```bash
go run ./cmd/server
```

The server is now running on `http://localhost:8080` !

## Tech Stack

- **Go** — server and web templates
- **PostgreSQL** — persistent storage
- **chi** — HTTP router ([go-chi/chi](https://github.com/go-chi/chi))
- **JWT** — authentication ([golang-jwt](https://github.com/golang-jwt/jwt))
- **argon2id** — password hashing ([alexedwards/argon2id](https://github.com/alexedwards/argon2id))
- **sqlc** — type-safe SQL ([sqlc](https://sqlc.dev))
- **goose** — database migrations ([pressly/goose](https://github.com/pressly/goose))
- **Docker** — containerization

## Contributing

Clone the repo and set up your .env (see Quick Start above)

```bash
# Run tests
go test ./...

# Start the server
go run ./cmd/server
```

Open a pull request to the `main` branch to add new features or fix issues.