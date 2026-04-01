#!/bin/sh
set -e

echo "Running migrations..."
goose -dir /app/sql/schema postgres "$DB_URL" up

echo "Starting server..."
exec ./server