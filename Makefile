.PHONY: all build run test bench accuracy clean

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

bench:
	@echo "==> Running benchmarks..."
	go test -bench=. -benchmem ./...

clean:
	@echo "==> Cleaning up..."
	rm -f $(APP_NAME)

accuracy:
	@echo "==> Running Accuracy Benchmarks"
	go test -v ./internal/orchestrator -run TestIntegratedEvaluation

