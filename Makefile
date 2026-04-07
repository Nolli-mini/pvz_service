.PHONY: generate up down build run clean

generate:
	oapi-codegen -package api -generate types api.yaml > internal/api/models.go
	oapi-codegen -package api -generate gin-server api.yaml > internal/api/server.go

build: generate
	go build -o bin/app ./cmd/...

run: generate
	go run ./cmd/main.go

up:
	docker-compose up -d

down:
	docker-compose down -v

clean:
	rm -f internal/api/models.go internal/api/server.go
	rm -rf bin/