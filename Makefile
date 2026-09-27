.PHONY: up down logs backend-test frontend-test migrate seed

up:
	docker compose up --build -d mysql redis
	@echo "MySQL :3306  Redis :6379 ready. Run: make migrate, then: make seed, then (cd backend && go run ./cmd/api)"

down:
	docker compose down

logs:
	docker compose logs -f

backend-test:
	cd backend && go test ./... -count=1

frontend-test:
	cd frontend && npm test

migrate:
	docker exec -i taskmanager-mysql mysql -h 127.0.0.1 -P 3306 -utaskuser -ptaskpass taskdb < backend/migrations/001_create_tasks.up.sql

seed:
	cd backend && DATABASE_URL="$$DATABASE_URL" go run ./cmd/seed
