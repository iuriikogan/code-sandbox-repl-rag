.PHONY: all build run test clean docker-build

APP_NAME = code-sandbox
MAIN_PATH = ./cmd/sandbox/main.go

all: test build

build:
	@echo "==> Building $(APP_NAME)..."
	go build -o $(APP_NAME) $(MAIN_PATH)

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

