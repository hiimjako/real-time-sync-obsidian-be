.PHONY: run
run:
	@go run cmd/main.go

.PHONY: cli
cli:
	@go run cmd/cli/* 2>cli.log

.PHONY: test
test:
	go test ./... -count 1 -race




