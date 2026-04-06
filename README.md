# Code Sandbox REPL RAG

## Project Overview
This project is a Go-based agentic Retrieval-Augmented Generation (RAG) simulation. It demonstrates how to build an Orchestrator agent using Google's Gemini models (`gemini-2.5-pro`) that can autonomously execute generated Python code to process massive unstructured context data.

The application establishes an inter-process communication (IPC) channel between Go and a spawned Python process using standard input/output streams and JSON. The Python process performs local chunking and semantic search (Cosine Similarity) by requesting vector embeddings from the Go host, which proxies requests to the `text-embedding-004` model. Furthermore, the Python script can spin up sub-agents (`gemini-2.5-flash`) via the Go host to perform targeted tasks on chunks of data.

## Key Technologies
- **Language**: Go (`go 1.25.0`)
- **External Execution**: Python 3 (via `os/exec` tool calling)
- **SDK**: Google Cloud Vertex AI SDK (`cloud.google.com/go/vertexai`)
- **Models Used**:
  - `gemini-2.5-pro` (Orchestrator)
  - `gemini-2.5-flash` (Sub-agent worker)
  - `text-embedding-004` (Semantic search / embeddings)

## Architecture Details
1. **Orchestrator Setup**: The Go app spins up an orchestrator with `gemini-2.5-pro`, passing it an `execute_python_script` tool.
2. **Context Passing**: An unstructured dummy dataset is written to a temporary file. The path is provided to the generated Python script via the `CONTEXT_FILE` environment variable.
3. **IPC Loop**: When the Python script runs, it interacts with the Go host by printing JSON messages (e.g., `{"type": "embed", "chunk": "..."}`) to `stdout`. The Go app decodes this, executes the respective GenAI API calls (embeddings or flash sub-agents), and writes the results back to the Python process via `stdin`.
4. **Synthesis**: Once Python computes the top RAG chunks or sub-agent outputs, it returns the final context to Go via a `{"type": "done"}` IPC message, allowing the Orchestrator to generate the final synthesized output.

## Prerequisites
- Go 1.25+
- Python 3 available in your system's PATH
- Google Cloud Project with Vertex AI API enabled
- Authenticated via Application Default Credentials (e.g., `gcloud auth application-default login`)
- Docker & Docker Compose (optional, for containerized execution)

## Environment Variables
Ensure the following environment variables are set before running the application natively:
- `GOOGLE_CLOUD_PROJECT`: Your Google Cloud Project ID.
- `GOOGLE_CLOUD_LOCATION`: Your Vertex AI location (defaults to `us-central1` if not set).

## Building and Running

This project provides a `Makefile` to simplify common operations.

### Run Locally

To run the project directly:
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
make run
```

To build a binary and run it:
```bash
export GOOGLE_CLOUD_PROJECT="your-project-id"
make build
./code-sandbox
```

### Run via Docker

To run the application in a Docker container using `docker-compose`:
```bash
# Build the image
make docker-build

# Start the container
make docker-up
```

### Testing

Tests are enforced to run with `-count=1` to prevent caching:
```bash
make test
```

## Development Conventions
- **Code Style**: Standard Go conventions (use `gofmt` and `goimports`).
- **Logging**: The project uses the standard library's structured logging (`log/slog`).
- **Dependencies**: Managed via Go modules (`go.mod`, `go.sum`).
- **Tool Calling**: Handled manually via native `os/exec` and standard input/output pipes rather than containerized runtimes, requiring careful handling of JSON marshaling and scanner buffers (up to 10MB allocated for IPC).

## Project Structure
- `cmd/sandbox/main.go`: Main application entry point.
- `internal/ai/`: Wrappers and clients for Google Cloud Vertex AI interactions.
- `internal/data/`: Data generation and context handling.
- `internal/ipc/`: Go-Python Inter-Process Communication logic.
- `internal/orchestrator/`: Primary agent orchestration and GenAI coordination loop.
- `internal/python/`: Python subprocess execution and management.
