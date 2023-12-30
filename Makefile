build:
	docker-compose build crm-system

run:
	docker-compose up crm-system

test:
	go test -v ./...

migrate:
	migrate -path ./db/migrations -database 'postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable' up

swag:
	swag init -g cmd/main.go