.PHONY: all build clean proto run-http run-grpc migrate test

BINARY_NAME=go-users
PROTO_DIR=api
MIGRATIONS_DIR=migrations
PROTO_OUT_DIR=internal/infrastructure/server/grpc/gen

all: proto build

build:
	@echo "Building binary..."
	go build -o $(BINARY_NAME) cmd/app/main.go

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -rf $(PROTO_OUT_DIR)

proto:
	@echo "Generating protobuf code..."
	mkdir -p $(PROTO_OUT_DIR)
	protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/v1/user.proto

run-http: build
	@echo "Running HTTP server..."
	./$(BINARY_NAME) http-server

run-grpc: build
	@echo "Running gRPC server..."
	./$(BINARY_NAME) grpc-server

migrate:
	@echo "Running migrations..."
	./$(BINARY_NAME) migrate

test:
	@echo "Running tests..."
	go test -v ./...

docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME) .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 50051:50051 $(BINARY_NAME) 