PHONY: shortening
shortening:
	@go run ./cmd/services/shortening

PHONY: up
up:
	@docker-compose up --build

PHONY: down
down:
	@docker-compose down

PHONY: downv
downv:
	@docker-compose down -v