.PHONY: fmt test test-cover test-short

fmt:
	@go fmt ./...

test:
	@go test -v ./...

test-cover:
	@go test -v -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

test-short:
	@go test -v -short ./...