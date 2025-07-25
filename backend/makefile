# Makefile

DB_PATH=records.db
SQLITE=sqlite3
MIGRATIONS=sql/init.sql

.PHONY: migrate clean run

## Run SQL migration file on the SQLite DB
migrate:
	@echo "Applying migration to $(DB_PATH)..."
	$(SQLITE) $(DB_PATH) < $(MIGRATIONS)
	@echo "Migration complete."

## Delete DB file
clean:
	@rm -f $(DB_PATH)
	@echo "Deleted $(DB_PATH)"

## Start the backend using air
run:
	air
