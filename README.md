# Code Sandbox REPL RAG

## Project Overview
This project is a Go-based agentic Retrieval-Augmented Generation (RAG) simulation. It demonstrates how to build an Orchestrator agent using Google's Gemini models (`gemini-3-flash-preview`) that can autonomously execute generated Python code to process massive unstructured context data.

The application leverages the **Vertex AI Agent Engine Code Execution Sandbox** (`cloud.google.com/go/aiplatform/apiv1beta1`) for secure, isolated Python execution. The Python script runs in the cloud and uses the `vertexai` Python SDK to perform local chunking, generate embeddings (`text-embedding-004`), and calculate Cosine Similarity. Furthermore, the Python script can spin up sub-agents (`gemini-3.1-flash-lite-preview`) within the sandbox to perform targeted tasks on chunks of data.

Alternatively, a `LocalRunner` is provided that executes Python locally via `os/exec` and communicates with the Go host over an IPC channel via JSON over standard input/output.

### Key Technologies
- **Language**: Go (`go 1.25.0`)
- **SDK**: Google GenAI SDK (`google.golang.org/genai`) & Vertex AI API (`cloud.google.com/go/aiplatform`)
- **Models Used**:
  - `gemini-2.5-flash` (Orchestrator)
  - `gemini-2.5-flash` (Sub-agent worker)
  - `gemini-2.5-pro` (Final Synthesis)
  - `text-embedding-004` (Semantic search / embeddings)
- **External Execution**: Vertex AI Agent Engine Code Execution Sandbox (Primary) / Local Python 3 (Fallback)

## Architecture Details
1. **Orchestrator Setup**: The Go app spins up an orchestrator with `gemini-2.5-flash`, passing it an `execute_python_script` tool.
2. **Context Passing**: An unstructured dummy dataset is created. The content is passed into the Vertex AI Sandbox via the API.
3. **Execution**: When the generated Python script runs in the sandbox, it interacts directly with Vertex AI using the injected `PROJECT_ID` and `LOCATION`. It chunks the data, gets embeddings, and performs similarity search locally within the sandbox.
4. **Synthesis**: Once Python computes the top RAG chunks or sub-agent outputs, it returns the final context to Go as a JSON-formatted standard output. The Orchestrator then generates the final synthesized output.

## Prerequisites
- Go 1.25+
- Google Cloud Project with Vertex AI API enabled
- Authenticated via Application Default Credentials (e.g., `gcloud auth application-default login`)
- Docker & Docker Compose (optional, for containerized execution)

### Environment Variables
You must set the following environment variable before running the application:
- `GOOGLE_CLOUD_PROJECT`: Your Google Cloud Project ID.

Note: The application uses the **`us-central1`** Vertex AI endpoint because the Agent Engine Code Execution Sandbox is currently only available in that region.

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

## Project Structure
- `cmd/sandbox/main.go`: Main application entry point.
- `internal/ai/`: Wrappers and clients for Google Cloud Vertex AI interactions.
- `internal/data/`: Data generation and context handling.
- `internal/ipc/`: Go-Python Inter-Process Communication logic.
- `internal/orchestrator/`: Primary agent orchestration and GenAI coordination loop.
- `internal/python/`: Python subprocess execution and management.
