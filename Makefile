PHONY: shortening
shortening:
	@go run ./cmd/services/shortening

PHONY: redirecting
redirecting:
	@go run ./cmd/services/redirecting

PHONY: up
up:
	@docker-compose up --build

PHONY: down
down:
	@docker-compose down

PHONY: downv
downv:
	@docker-compose down -v

PHONY: test_url_mapping_repo
test_url_mapping_repo:
	@go test -race -count=1 ./internal/repos/urlmappingrepo
