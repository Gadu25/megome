build:
	@go build -o bin/megome cmd/main.go

test:
	@go test -v ./...

run: build
	@./bin/megome