# Code Sandbox REPL RAG

## Project Overview
This project is a Go-based agentic Retrieval-Augmented Generation (RAG) simulation. It demonstrates how to build an Orchestrator agent using Google's latest Gemini models (`gemini-3.1-flash-lite-preview`) that can autonomously execute generated Python code to process massive unstructured context data using mathematical tools like BM25 and Standard Deviation.

The application establishes an inter-process communication (IPC) channel between Go and a spawned Python process using standard input/output streams and JSON. The Python process performs local chunking and semantic search (Cosine Similarity) by requesting vector embeddings from the Go host, which proxies requests to the `text-embedding-004` (or AI Studio's `gemini-embedding-001`) model. Furthermore, the Python script uses this IPC to spin up concurrent sub-agents (`gemini-2.5-flash`) via the Go host to perform targeted Map-Reduce tasks on chunks of data.

### Key Technologies
- **Language**: Go (`go 1.25.0`)
- **SDK**: Google GenAI SDK (`google.golang.org/genai`)
- **Models Used**:
  - `gemini-3.1-flash-lite-preview` (Orchestrator)
  - `gemini-2.5-flash` (Sub-agent Map-Reduce Swarm worker)
  - `gemini-2.5-pro` (Final Synthesis)
  - `text-embedding-004` / `gemini-embedding-001` (Semantic search / embeddings)
- **External Execution**: Local Python 3 subprocess with a 50MB+ capacity `bufio.Reader` standard I/O IPC.

## Architecture Details
1. **Orchestrator Setup**: The Go app spins up an orchestrator with `gemini-3.1-flash-lite-preview`, passing it the `run_rag_agent_logic` tool and injecting a suite of powerful native Python analytics tools (BM25, Rocchio expansion, Standard Deviation outliers).
2. **IPC Loop**: When the Python script runs, it interacts with the Go host by printing JSON messages (e.g., `{"type": "batch_embed", "chunks": ["..."]}`) to `stdout`. The Go app parses this, executes the respective concurrent GenAI API calls (embeddings or flash sub-agents) using a rate-limiting 10-worker semaphore, and writes the results back to the Python process via `stdin`.
3. **Synthesis**: Once Python recursively traces the multi-hop reasoning over the 150K+ context chunks, it returns the final, highly compressed clues to Go via a `{"type": "done"}` IPC message, allowing the Orchestrator to generate the final synthesized output via the `submit_clues_for_synthesis` function call.

## Building and Running

### Prerequisites
- Go 1.25+
- Python 3 available in your system's PATH.
- Google Cloud Project with Vertex AI API enabled OR an AI Studio `GEMINI_API_KEY`.
- Authenticated via Application Default Credentials (e.g., `gcloud auth application-default login`).

### Environment Variables
You must set the following environment variable before running the application:
- `GEMINI_API_KEY` (Strongly Recommended): Your Gemini Developer API Key. Using this gives access to the latest `gemini-3.1-flash-lite-preview` orchestration model.
- `GOOGLE_CLOUD_PROJECT`: Your Google Cloud Project ID (Fallback for Vertex AI).
- `GOOGLE_CLOUD_LOCATION`: Your Vertex AI target region. Defaults to `us-central1`.

### Commands
To build and run the application:
```bash
# Run directly with Developer API
export GEMINI_API_KEY="your-key-here"
go run cmd/sandbox/main.go

# Or test accuracy
make accuracy
```

## Development Conventions
- **Code Style**: Standard Go conventions (use `gofmt` and `goimports`).
- **Logging**: The project uses the standard library's structured logging (`log/slog`).
- **Dependencies**: Managed via Go modules (`go.mod`, `go.sum`).
- **Tool Calling**: Handled manually via native `os/exec` and standard input/output pipes. To ensure stable parsing of JSON strings up to 50MB between processes, standard library buffers and strict Python string-escapes are prioritized.