# Makefile

DB_PATH=records.db
SQLITE=sqlite3
MIGRATIONS=sql/init.sql
MIGRATIONS_DOWN=sql/dropData.sql

.PHONY: migrate clean run

## Run SQL migration file on the SQLite DB
migrationup:
	make clean
	@echo "Applying migration to $(DB_PATH)..."
	$(SQLITE) $(DB_PATH) < $(MIGRATIONS)
	@echo "Migration complete."

# Run SQL migration file on the SQLite DB
migrationdown:
	@echo "Reverting migration on $(DB_PATH)..."
	$(SQLITE) $(DB_PATH) < $(MIGRATIONS_DOWN)
	@echo "Data in db dropped."

## Delete DB file
clean:
	@echo "Deleting $(DB_PATH)..."
	@rm -f $(DB_PATH)
	@echo "Deleted $(DB_PATH)"

## Start the backend using air
run:
	air
