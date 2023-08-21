.PHONY: all
all:
	go build ./...
	go test -cover -coverprofile=coverage.out ./...
	go vet ./...
	staticcheck -checks all ./...
