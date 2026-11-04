.PHONY: build test generate generate-check migrate-up migrate-down

build:
	go build -buildvcs=false ./...

test:
	go test ./...

generate:
	go tool ogen -clean -package api -target internal/adapter/primary/http/api -config codegen/ogen.yml api/openapi.yml
	rm -f internal/adapter/secondary/postgres/sqlc/*.sql.go
	go tool sqlc generate -f codegen/sqlc.yml
	go tool wire gen ./internal/bootstrap

generate-check:
	./scripts/generate-check.sh

migrate-up:
	go run ./cmd/wahrwelt migrate

migrate-down:
	go run ./cmd/wahrwelt migrate-down
