build:
	@go build -o bin/megome cmd/main.go

test:
	@go test -v ./...

run: build
	@./bin/megome

migration:
	@migrate create -ext sql -dir cmd/migrate/migrations $(name)

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down

seed:
	@go run cmd/seed/main.go