up:
	docker compose up -d

down:
	docker compose down

restart: down up

test:
	@cd db/sqlc && go test

server:
	go run main.go

tests:
	go test -v ./...

.PHONY: tests up down restart