BINARY_NAME=main.out

build:
	@go build -o bin/level cmd/main.go

test:
	@go test -v ./...

run: build
	@./bin/level
