PHONY: build_shortening
build_shortening:
	@go build -o ./bin/services/shortening ./cmd/services/shortening

PHONY: shortening
shortening: build_shortening
	@./bin/services/shortening -env-file=.env.local

PHONY: build_redirecting
build_redirecting:
	@go build -o ./bin/services/redirecting ./cmd/services/redirecting

PHONY: redirecting
redirecting:
	@./bin/services/redirecting -env-file=.env.local

PHONY: build_analytics
build_analytics:
	@go build -o ./bin/services/analytics ./cmd/services/analytics

PHONY: analytics
analytics:
	@./bin/services/analytics -env-file=.env.local

PHONY: build_clicks
build_clicks:
	@go build -o ./bin/workers/clicks ./cmd/workers/clicks

PHONY: clicks
clicks:
	@./bin/workers/clicks -env-file=.env.local

PHONY: build_clicksbatcher
build_clicksbatcher:
	@go build -o ./bin/workers/clicksbatcher ./cmd/workers/clicksbatcher

PHONY: clicksbatcher
clicksbatcher:
	@./bin/workers/clicksbatcher -env-file=.env.local

PHONY: build_cassandra
build_cassandra:
	@go build -o ./bin/migrators/cassandra ./cmd/migrators/cassandra

PHONY: cassandra
cassandra:
	@./bin/migrators/cassandra -env-file=.env.local

PHONY: build_clickhouse
build_clickhouse:
	@go build -o ./bin/migrators/clickhouse ./cmd/migrators/clickhouse

PHONY: clickhouse
clickhouse:
	@./cmd/migrators/clickhouse -env-file=.env.local

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
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmapping

PHONY: test_analyticclicks_repo
test_analyticclicks_repo:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/analyticclicks

PHONY: test_redirecting
test_redirecting:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmapping ./internal/core/services/redirecting

PHONY: test_shortening
test_shortening:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmapping ./internal/core/services/shortening

PHONY: test_analytics
test_analytics:
	@CGO_ENABLED=1 go test -race -count=6 ./internal/repos/analyticclicks ./internal/core/services/analytics

PHONY: tests
tests:
	@CGO_ENABLED=1 go test -race -count=1 ./internal/repos/urlmapping ./internal/repos/analyticclicks ./internal/core/services/redirecting ./internal/core/services/shortening ./internal/core/services/analytics