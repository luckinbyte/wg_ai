.PHONY: all build test clean proto

all: proto build

build:
	go build -o bin/game ./cmd/game
	go build -o bin/login ./cmd/login
	go build -o bin/db ./cmd/db

test:
	go test -v ./...

clean:
	rm -rf bin/

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/cs/*.proto proto/ss/*.proto
