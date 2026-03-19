.PHONY: dev build migrate clean test install

# Default target
all: build

# Install dependencies
install:
	@echo "📦 Installing Go dependencies..."
	go mod download
	@echo "📦 Installing frontend dependencies..."
	cd frontend && npm install

# Development mode (runs backend with hot reload)
dev:
	@echo "🚀 Starting development server..."
	@echo "Backend will run on http://localhost:8080"
	@echo "Frontend dev server will run on http://localhost:5173"
	@echo ""
	@echo "In another terminal, run: cd frontend && npm run dev"
	@echo ""
	go run cmd/server/main.go

# Build everything
build: build-frontend build-backend

# Build backend binary
build-backend:
	@echo "🔨 Building backend..."
	@$(eval VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || git rev-parse --short HEAD))
	@$(eval COMMIT := $(shell git rev-parse --short HEAD))
	@$(eval BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S'))
	go build -ldflags "-X github.com/kayushkin/inber-party/internal/version.Version=$(VERSION) \
		-X github.com/kayushkin/inber-party/internal/version.GitCommit=$(COMMIT) \
		-X github.com/kayushkin/inber-party/internal/version.BuildTime=$(BUILD_TIME)" \
		-o bin/inber-party cmd/server/main.go
	@echo "✅ Backend built: bin/inber-party (version: $(VERSION))"

# Build frontend for production
build-frontend:
	@echo "🔨 Building frontend..."
	cd frontend && npm run build
	@echo "✅ Frontend built: frontend/dist/"

# Run migrations (database must be running)
migrate:
	@echo "🔄 Migrations run automatically on server start"
	@echo "To run server: make dev or ./bin/inber-party"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf frontend/dist/
	@echo "✅ Clean complete"

# Run tests
test:
	@echo "🧪 Running Go tests..."
	go test ./...
	@echo "🧪 Running frontend tests..."
	cd frontend && npm test

# Run production build
run: build
	@echo "🚀 Starting production server..."
	./bin/inber-party

# Help
help:
	@echo "Inber Party - Available commands:"
	@echo ""
	@echo "  make install       - Install all dependencies"
	@echo "  make dev          - Run backend in development mode"
	@echo "  make build        - Build both frontend and backend"
	@echo "  make build-backend - Build only the backend"
	@echo "  make build-frontend - Build only the frontend"
	@echo "  make run          - Build and run in production mode"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run all tests"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  DATABASE_URL - PostgreSQL connection string"
	@echo "                 (default: postgres://localhost:5432/inber_party?sslmode=disable)"
	@echo "  PORT         - Server port (default: 8080)"
