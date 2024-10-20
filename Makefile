.PHONY: run
run:
	@go run cmd/main.go

.PHONY: cli
cli:
	@go run cmd/cli/* 2>cli.log

.PHONY: test
test:
	go test ./... -count 1 -race

.PHONY: lint
lint:
	golangci-lint run --fix
	npx prettier-pnp --pnp prettier-plugin-sql --write migrations/*.sql --config .github/config/.prettierrc





