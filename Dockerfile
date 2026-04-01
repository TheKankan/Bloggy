# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Install goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .
RUN go build -o server ./cmd/server

# Run stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/sql/schema ./sql/schema
COPY --from=builder /app/web ./web
COPY --from=builder /app/scripts/entrypoint.sh .

RUN chmod +x entrypoint.sh

EXPOSE 8080
CMD ["./entrypoint.sh"]