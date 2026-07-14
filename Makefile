PHONY: shortening
shortening:
	@go run ./cmd/services/shortening

PHONY: redirecting
redirecting:
	@go run ./cmd/services/redirecting

PHONY: analytics
analytics:
	@go run ./cmd/services/analytics

PHONY: clickscollector
clicks:
	@go run ./cmd/workers/clickscollector

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
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmappingrepo

PHONY: test_analytic_repo
test_analytic_repo:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/analyticrepo

PHONY: test_redirecting
test_redirecting:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmappingrepo ./internal/core/services/redirecting

PHONY: test_shortening
test_shortening:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmappingrepo ./internal/core/services/shortening

PHONY: test_analytics
test_analytics:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/analyticrepo ./internal/core/services/analytics

PHONY: tests
tests:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmappingrepo ./internal/core/services/redirecting ./internal/core/services/shortening