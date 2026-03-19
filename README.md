![CI](https://github.com/TheKankan/Bloggy/actions/workflows/ci.yml/badge.svg)

# Bloggy

A blog coded in go allowing users to register, create posts and comment other users posts

## Contributing

1. Clone the repo
```bash
git clone https://github.com/TheKankan/Terminal_Chat.git
```

2. Edit .env.example : rename it .env and change the values inside it

3. Run migrations
```bash
goose -dir sql/schema postgres "$DB_URL" up
```

Open a pull request to the `main` branch to add new features or fix issues.