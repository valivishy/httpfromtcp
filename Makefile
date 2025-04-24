.PHONY: build run test clean lint build-and-run

lint:
	golangci-lint run ./...

test:
	go test -v ./...

lint-test: lint test