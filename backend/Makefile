.PHONY: dev build test migrate migrate-down sqlc lint docker-up docker-down

# Start local infrastructure (postgres + redis)
docker-up:
	docker compose up -d postgres redis

# Stop local infrastructure
docker-down:
	docker compose down

# Run all database migrations
migrate:
	migrate -path migrations/ -database "$${DATABASE_URL}" up

# Roll back all migrations
migrate-down:
	migrate -path migrations/ -database "$${DATABASE_URL}" down

# Run migrations and seed data
migrate-seed: migrate
	psql "$${DATABASE_URL}" -f migrations/seed/scenarios_seed.sql

# Generate sqlc type-safe code
sqlc:
	sqlc generate

# Build the Go binary
build:
	go build -o bin/server ./cmd/server

# Run the server locally
dev:
	go run ./cmd/server

# Run all tests
test:
	go test ./... -v -timeout 60s

# Run tests with race detector
test-race:
	go test ./... -race -timeout 120s

# Run linter
lint:
	golangci-lint run ./...

# Generate a new RSA key pair for JWT
gen-keys:
	mkdir -p secrets
	openssl genrsa -out secrets/jwt_private.pem 2048
	openssl rsa -in secrets/jwt_private.pem -pubout -out secrets/jwt_public.pem
	echo "Keys generated in secrets/"

# Tidy go modules
tidy:
	go mod tidy
