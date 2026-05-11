.PHONY: all build run test clean docker-build docker-up docker-down build-worker

APP_NAME = code-sandbox
WORKER_IMAGE = python-worker
MAIN_PATH = ./cmd/sandbox/main.go

all: test build

build:
	@echo "==> Building $(APP_NAME)..."
	go build -o $(APP_NAME) $(MAIN_PATH)

build-worker:
	@echo "==> Building $(WORKER_IMAGE)..."
	docker build -t $(WORKER_IMAGE):latest ./deploy/worker

run:
	@echo "==> Running $(APP_NAME)..."
	go run $(MAIN_PATH)

test:
	@echo "==> Running tests..."
	go test -count=1 ./...

clean:
	@echo "==> Cleaning up..."
	rm -f $(APP_NAME)

docker-build:
	@echo "==> Building Docker image..."
	docker-compose build

docker-up:
	@echo "==> Starting Docker container..."
	docker-compose up

docker-down:
	@echo "==> Stopping Docker container..."
	docker-compose down
