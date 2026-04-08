# Code Sandbox REPL RAG

## Project Overview
*Secure, Isolated Execution* By leveraging the Vertex AI Agent Engine Code Execution Sandbox, the system safely runs AI-generated Python code in a secure cloud environment, preventing malicious or flawed code from affecting the host machine.

*Dynamic Data Processing (Agentic RAG)* Instead of relying on a static data pipeline, an orchestrator agent autonomously writes custom code to handle chunking, embedding generation, and cosine similarity searches specifically tailored to the immediate dataset.

*Hierarchical Agent Swarms* The executed Python scripts can dynamically spin up smaller sub-agents within the sandbox to perform highly targeted tasks on specific data chunks in parallel.

*Strategic Model Routing* The architecture optimizes performance and cost by using faster models (Gemini Flash) for orchestration and sub-agent work, a specialized embedding model for semantic search, and a heavier reasoning model (Gemini Pro) for final synthesis.

### Benefits

*Uncompromised Security for Code Generation*: Because LLMs can sometimes hallucinate destructive commands, sandboxing the execution ensures enterprise-grade security, allowing developers to safely experiment with code-generating agents.

*Massive Scalability*: Shifting the heavy lifting—chunking, embedding, and similarity search—into an isolated cloud environment allows the application to process massive unstructured datasets that would otherwise overwhelm local memory.

*Highly Contextual Retrieval*: Traditional RAG relies on rigid, predefined chunking strategies that often miss context. This agentic approach writes bespoke logic to navigate the data, drastically improving the relevance of the retrieved context.

*Superior Final Output*: By ensuring that the data is meticulously filtered and processed by sub-agents before being handed to a powerful synthesis model, the final generation achieves a higher degree of accuracy and reasoning quality.

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
