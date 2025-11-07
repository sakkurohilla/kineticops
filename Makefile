.PHONY: up down migrate build-backend

up:
	docker-compose up -d

down:
	docker-compose down

migrate:
	docker-compose run --rm migrator

build-backend:
	cd backend && go build -o kineticops-backend ./cmd/server
