.PHONY: clear backend objectstorage database all down up restart

clear:
	@clear

backend:
	@echo "Rebuilding backend..."
	@cd backend/app && make generate build
	@cd backend && ./build.sh
	@echo "Updating backend container..."
	@docker compose up -d --no-deps --force-recreate backend

objectstorage:
	@echo "Rebuilding object storage..."
	@cd objectStorage/ && ./build.sh
	@echo "Updating object storage container..."
	@docker compose up -d --no-deps --force-recreate objectstorage

database:
	@echo "Rebuilding database..."
	@cd db/ && ./build.sh
	@echo "Updating database container..."
	@docker compose up -d --no-deps --force-recreate database

all: backend database objectstorage

down:
	@docker compose down -v --remove-orphans

up:
	@docker compose up -d --force-recreate

restart: down up
