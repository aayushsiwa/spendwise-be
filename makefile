# --- Configurable Variables ---
DB_PATH           := records.db
SQLITE            := sqlite3
MIGRATIONS        := sql/init.sql
MIGRATIONS_DOWN   := sql/dropData.sql

BINARY_NAME       := main
IMAGE_NAME        := expense-backend
CONTAINER_PORT    := 8080
HOST_PORT         := 8080

MIN_COVERAGE := 56

# --- Targets ---
.PHONY: all migrationup migrationdown clean dev build docker run docker-clean push coverage help

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

## Run tests with coverage
coverage: ## Run tests with coverage and enforce minimum coverage
	@echo "Running tests with coverage..."
	@mkdir -p tmp
	@go test -coverprofile=tmp/coverage.out -covermode=count ./...

	@echo ""
	@echo "--- Per-function coverage ---"
	@go tool cover -func=tmp/coverage.out

	@coverage=$$(go tool cover -func=tmp/coverage.out | awk '/total:/ {print substr($$3, 1, length($$3)-1)}'); \
	echo ""; \
	echo "Total coverage: $$coverage%"; \
	if awk 'BEGIN { exit !('"$$coverage"' >= $(MIN_COVERAGE)) }'; then \
		echo "✅ Coverage meets threshold ($(MIN_COVERAGE)%)"; \
	else \
		echo "❌ Coverage ($$coverage%) is below threshold ($(MIN_COVERAGE)%)"; \
		exit 1; \
	fi

	@go tool cover -html=tmp/coverage.out -o tmp/coverage.html
	@echo "📄 HTML report: tmp/coverage.html"

## Run tests with coverage and open HTML report
coverage-html: ## Generate HTML coverage report
	@echo "Running tests with coverage..."
	@mkdir -p tmp
	go test -coverprofile=tmp/coverage.out -covermode=count ./...
	go tool cover -html=tmp/coverage.out -o tmp/coverage.html
	@echo "✅ HTML coverage report generated: tmp/coverage.html"

	@# Open report automatically (Linux/macOS)
	@if command -v xdg-open >/dev/null 2>&1; then \
		xdg-open tmp/coverage.html; \
	elif command -v open >/dev/null 2>&1; then \
		open tmp/coverage.html; \
	else \
		echo "Open tmp/coverage.html in your browser"; \
	fi

coverage-check: ## Fail if coverage is below threshold
	@mkdir -p tmp
	@go test -coverprofile=tmp/coverage.out -covermode=count ./...
	@coverage=$$(go tool cover -func=tmp/coverage.out | awk '/total:/ {sub("%","",$$3); print $$3}'); \
	if awk 'BEGIN { exit !('"$$coverage"' < $(MIN_COVERAGE)) }'; then \
		echo "❌ Coverage $$coverage% is below $(MIN_COVERAGE)%"; \
		exit 1; \
	elif awk 'BEGIN { exit !('"$$coverage"' > $(MIN_COVERAGE)) }'; then \
		echo "✅ Coverage is greater than $(MIN_COVERAGE)%: $$coverage%"; \
		echo "Suggest increasing MIN_COVERAGE to $$coverage%"; \
	else \
		echo "✅ Coverage: $$coverage%"; \
	fi

## Run Go modernize
modernize: ## Run Go modernize
	@echo "Running Go modernize..."
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix ./...
	@echo "✅ Modernize passed"

lint: ## Run Go linter
	@echo "Running Go linter..."
	 golangci-lint run --fix ./...
	@echo "✅ Linter passed"

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
pre-push: modernize lint check test coverage-check ## Run all pre-push checks (fmt, vet, test)
	@echo "✅ Pre-push checks passed"

## Show help for all commands
help:
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
