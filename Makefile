PHONY: shortening
shortening:
	@go run ./cmd/services/shortening -env-file=.env.local

PHONY: redirecting
redirecting:
	@go run ./cmd/services/redirecting -env-file=.env.local

PHONY: analytics
analytics:
	@go run ./cmd/services/analytics -env-file=.env.local

PHONY: clicks
clicks:
	@go run ./cmd/workers/clicks -env-file=.env.local

PHONY: postgres
postgres:
	@go run ./cmd/migrators/postgres -env-file=.env.local

PHONY: cassandra
cassandra:
	@go run ./cmd/migrators/cassandra -env-file=.env.local

PHONY: clickhouse
clickhouse:
	@go run ./cmd/migrators/clickhouse -env-file=.env.local

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
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmappingsrepo

PHONY: test_analytics_repo
test_analytics_repo:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/analyticsrepo

PHONY: test_redirecting
test_redirecting:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmappings ./internal/core/services/redirecting

PHONY: test_shortening
test_shortening:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmappings ./internal/core/services/shortening

PHONY: test_analytics
test_analytics:
	@CGO_ENABLED=1 go test -race -count=6 ./internal/repos/analyticsrepo ./internal/core/services/analytics

PHONY: tests
tests:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmappingsrepo ./internal/repos/analyticsrepo ./internal/core/services/redirecting ./internal/core/services/shortening ./internal/core/services/analytics