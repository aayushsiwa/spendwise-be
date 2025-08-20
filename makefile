# --- Configurable Variables ---
DB_PATH           := records.db
SQLITE            := sqlite3
MIGRATIONS        := sql/init.sql
MIGRATIONS_DOWN   := sql/dropData.sql

BINARY_NAME       := main
IMAGE_NAME        := expense-backend
CONTAINER_PORT    := 8080
HOST_PORT         := 8080

# --- Targets ---
.PHONY: all migrationup migrationdown clean dev build docker run docker-clean push help

## Run full setup: clean, migrate, build docker image, and run container
all: run

## Apply migration SQL to the SQLite database
migrationup: clean ## Remove the older db and apply migration
	@echo "Applying migration to $(DB_PATH)..."
	$(SQLITE) $(DB_PATH) < $(MIGRATIONS)
	@echo "Migration applied."

## Revert migration (drop tables/data)
migrationdown: ## Revert the migration and drop data from db
	@echo "Reverting migration on $(DB_PATH)..."
	$(SQLITE) $(DB_PATH) < $(MIGRATIONS_DOWN)
	@echo "Migration reverted (data dropped)."

## Remove the SQLite database file
clean: ## Remove the SQLite db file
	@echo "Removing $(DB_PATH) $(DB_PATH)-shm $(DB_PATH)-wal ..."
	@if rm -f $(DB_PATH) $(DB_PATH)-shm $(DB_PATH)-wal; then \
		echo "Deleted $(DB_PATH)."; \
	else \
		echo "Failed to remove database files"; \
	fi


## Start the development server with Air for live reloading
dev: ## Start the development server with Air for live reloading
	@echo "Starting development server..."
	@command -v air >/dev/null 2>&1 && air || (echo "Air not found. Falling back to: go run ."; go run .)

## Build Go binary for Linux (Docker-compatible)
build: ## Build the application
	@echo "Building Go binary..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_NAME) .
	@echo "Binary built: $(BINARY_NAME)"

## Build Docker image
docker: build ## Build the Docker image for the application
	@echo "Building Docker image: $(IMAGE_NAME)..."
	docker build -t $(IMAGE_NAME) .
	@echo "Docker image built: $(IMAGE_NAME)"
	$(MAKE) docker-clean

## Remove binary after Docker build
docker-clean: ## Clean up the binary after building Docker image
	@echo "Cleaning up binary..."
	@rm -f $(BINARY_NAME)
	@echo "Binary removed."

## Run the Docker container (make sure DB file exists)
run: docker  ## Run the application
	@echo "Running Docker container on port $(HOST_PORT)..."
	@echo "Mounting $(DB_PATH)..."
	docker run --rm -p $(HOST_PORT):$(CONTAINER_PORT) \
		-v $(PWD)/$(DB_PATH):/root/$(DB_PATH) \
		$(IMAGE_NAME)

## Push Docker image to Docker Hub (optional)
push: ## Push the Docker image to Docker Hub
	@echo "Pushing Docker image to Docker Hub..."
	docker push $(IMAGE_NAME)
	@echo "Image pushed: $(IMAGE_NAME)"

## Run Go tests
test: ## Run Go unit tests
	@echo "Running tests..."
	go test ./...
	@echo "✅ Tests passed"

## Run go vet and go fmt
check: ## Run go vet and go fmt
	@echo "Running go fmt..."
	@unformatted=$$(go fmt ./...); \
	if [ -n "$$unformatted" ]; then \
		echo "❌ Unformatted files found:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi
	@echo "Running go vet..."
	go vet ./...
	@echo "✅ Formatting and vetting passed"

## Run all checks before push
pre-push: check test ## Run all pre-push checks (fmt, vet, test)
	@echo "✅ Pre-push checks passed"

## Show help for all commands
help:
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""